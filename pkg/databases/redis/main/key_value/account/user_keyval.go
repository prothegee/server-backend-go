package db_rd_main_account_user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"showcase-backend-go/pkg"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// --------------------------------------------------------- //

// account kv of user session holder type
type UserSession struct{}

// db_rd_main user session type
type UserSession_t struct {
	Id         uuid.UUID
	Dt_Created time.Time
	Dt_Expired time.Time
}

// db_rd_main user session type json
type UserSession_tj struct {
	Id         uuid.UUID `json:"id"`
	Dt_Created time.Time `json:"dt_created"`
	Dt_Expired time.Time `json:"dt_expired"`
}

// conversion UserSession_t to UserSession_tj
//
// receiver: d UserSession_t
//
// return: UserSession_tj
func (d UserSession_t) ToJSON() UserSession_tj {
	return UserSession_tj{
		Id:         d.Id,
		Dt_Created: d.Dt_Created,
		Dt_Expired: d.Dt_Expired,
	}
}

// --------------------------------------------------------- //

const (
	// %[1]s = must existing user id
	NS_ACCOUNT_USER_ID = "account:user:%[1]s"
)

const (
	UserSessionKEY_session = "session"
)

const (
	UserSessionSESSION_id         = "id"
	UserSessionSESSION_dt_created = "dt_created"
	UserSessionSESSION_dt_expired = "dt_expired"
)

// --------------------------------------------------------- //

// create new session id from existing userId
//
// has default ttl for 6 minutes
//
// params:
// 	rdb *redis.Client - must db_rd.MainDb
// 	ctx context.Context
// 	userId uuid.UUID
//
// return: error
func (_ UserSession) SetNewSession(rdb *redis.Client, ctx context.Context,
	userId uuid.UUID) error {
	key := fmt.Sprintf(NS_ACCOUNT_USER_ID, userId.String())

	id, err := pkg.GenerateUUID(pkg.UUID_V7)
	if err != nil {
		return err
	}
	dtCreated := time.Now()
	dtExpired := dtCreated.Add(time.Minute * 6)

	sessionData := UserSession_t{
		Id:         id,
		Dt_Created: dtCreated,
		Dt_Expired: dtExpired,
	}

	jsonBytes, err := json.Marshal(sessionData.ToJSON())
	if err != nil {
		return err
	}

	err = rdb.HSet(ctx, key, UserSessionKEY_session, string(jsonBytes)).Err()
	if err != nil {
		return err
	}

	ttl := time.Until(dtExpired)
	err = rdb.Expire(ctx, key, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed set TTL for session: %w", err)
	}

	return nil
}

// get existing session data from userid
//
// params:
// 	rdb *redis.Client - must db_rd.MainDb
// 	ctx context.Context
// 	userId uuid.UUID
//
// return: (UserSession_tj, error) - (data, nil if ok)
func (_ UserSession) GetSessionData(rdb *redis.Client, ctx context.Context,
	userId uuid.UUID) (UserSession_tj, error) {
	var (
		res UserSession_tj
	)

	key := fmt.Sprintf(NS_ACCOUNT_USER_ID, userId.String())

	val, err := rdb.HGet(ctx, key, UserSessionKEY_session).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return res, errors.New("session not found")
		}
		// could be connection or else
		return res, fmt.Errorf("failed to get session from redis: %w", err)
	}

	if err := json.Unmarshal([]byte(val), &res); err != nil {
		return res, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return res, nil
}

// get existing session from userid
//
// params:
// 	rdb *redis.Client - must db_rd.MainDb
// 	ctx context.Context
// 	userId uuid.UUID
//
// return: (bool, error) - true if exists
func (_ UserSession) GetSessionExistence(rdb *redis.Client, ctx context.Context,
	userId uuid.UUID) (bool, error) {
	key := fmt.Sprintf(NS_ACCOUNT_USER_ID, userId.String())

	return rdb.HExists(ctx, key, UserSessionKEY_session).Result()
}

// delete existing session from userid
//
// params:
// 	rdb *redis.Client - must db_rd.MainDb
// 	ctx context.Context
// 	userId uuid.UUID
//
// return: (int64, error) - greater than 0 mean ok
func (_ UserSession) DeleteSession(rdb *redis.Client, ctx context.Context,
	userId uuid.UUID) (int64, error) {
	key := fmt.Sprintf(NS_ACCOUNT_USER_ID, userId.String())

	return rdb.HDel(ctx, key, UserSessionKEY_session).Result()
}
