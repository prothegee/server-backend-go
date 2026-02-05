/*
IMPORTANT:
- should be on last test
*/
package test_unittest

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	backend_api "showcase-backend-go/cmd/backend_api/api"
	backend_api_account "showcase-backend-go/cmd/backend_api/api/account"
	backend_api_auth "showcase-backend-go/cmd/backend_api/api/auth"
	backend_api_game1 "showcase-backend-go/cmd/backend_api/api/game1"
	"showcase-backend-go/pkg"
	mw "showcase-backend-go/pkg/middleware"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// --------------------------------------------------------- //

// net global backend_api test data
//
// cfg - server config
//
// server - as in target from backend listener
var (
	cfg    pkg.ConfigServer
	server string
)

// end-user temporal global backend_api test data
//
// userId
//
// email
//
// password
//
// authorizationData
//
// stashName
var (
	userId                                        uuid.UUID
	email, password, authorizationData, stashName string
)

// test server management
var (
	serverProcess *exec.Cmd
	testBinary    = "backend_api_test"
	projectRoot   string
	testDir       string
)

// --------------------------------------------------------- //

// find project root directory by looking for go.mod
func findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// navigate up until we find go.mod
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			break // reached root
		}
		wd = parent
	}

	return "", fmt.Errorf("project root not found")
}

// create database if not exists, mimics server's InitPgDbMain logic
// uses t.Log for proper test logging
func createDatabaseIfNotExists(t *testing.T, configPath string) {
	t.Helper()

	ctx := context.Background()
	content, err := pkg.ConfigServerLoad(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// extract postgres connection info
	pgConfig := content.Database.PostgreSQL.Main
	t.Logf("database config: host=%s, port=%d, user=%s, database=%s",
		pgConfig.Host, pgConfig.Port, pgConfig.User, pgConfig.Database)

	// build connection string without database name (connect to default 'postgres')
	var connStrBuilder strings.Builder
	connStrBuilder.WriteString("user=")
	connStrBuilder.WriteString(pgConfig.User)

	if len(pgConfig.Password) > 0 {
		connStrBuilder.WriteString(" password=")
		connStrBuilder.WriteString(pgConfig.Password)
	}

	connStrBuilder.WriteString(" host=")
	connStrBuilder.WriteString(pgConfig.Host)

	connStrBuilder.WriteString(" port=")
	connStrBuilder.WriteString(fmt.Sprintf("%d", pgConfig.Port))

	connStrBuilder.WriteString(" dbname=postgres") // connect to default database
	connStrBuilder.WriteString(" sslmode=")
	connStrBuilder.WriteString(pgConfig.SslMode)

	connStr := connStrBuilder.String()
	t.Logf("connecting to postgres with: %s", connStr)

	// connect to postgres database
	db, err := pgx.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close(ctx)

	t.Log("NOTICE: start L139")

	// create database if not exists
	// escape database name to prevent sql injection
	escapedDB := strings.ReplaceAll(pgConfig.Database, "'", "''")
	createDBSQL := fmt.Sprintf("CREATE DATABASE \"%s\"", escapedDB)

	t.Log("NOTICE: reaching L144")

	_, err = db.Exec(ctx, createDBSQL)
	if err != nil {
		// check if error is "database already exists" (sqlstate 42p04)
		if strings.Contains(err.Error(), "42P04") {
			t.Logf("database \"%s\" already exists, skipping creation", pgConfig.Database)
			return
		}
		t.Fatalf("failed to create database: %v", err)
	}

	t.Logf("database \"%s\" created successfully", pgConfig.Database)
}

// build and start backend server for integration tests
func setupTestServer(t *testing.T) {
	t.Helper()

	// find project root
	var err error
	projectRoot, err = findProjectRoot()
	if err != nil {
		t.Fatalf("failed to find project root: %v", err)
	}

	// get current test directory (where test is running from)
	testDir, err = os.Getwd()
	if err != nil {
		t.Fatalf("failed to get test directory: %v", err)
	}

	// load config to get database info
	configPath := filepath.Join(projectRoot, "config.json")
	t.Logf("loading config from: %s", configPath)

	// create database before building server
	createDatabaseIfNotExists(t, configPath)

	// build the server binary in project root
	t.Log("building server binary...")
	buildCmd := exec.Command("go", "build", "-o", testBinary, "./cmd/backend_api")
	buildCmd.Dir = projectRoot

	// capture output for debugging
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Logf("build output:\n%s", string(buildOutput))
		t.Fatalf("failed to build server: %v", err)
	}
	t.Log("server binary built successfully")

	// start the server with correct working directory
	serverProcess = exec.Command(filepath.Join(projectRoot, testBinary))
	serverProcess.Dir = testDir

	// redirect output for debugging
	serverProcess.Stdout = os.Stdout
	serverProcess.Stderr = os.Stderr

	t.Log("starting server...")
	if err := serverProcess.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	// load config using absolute path
	cfg, err = pkg.ConfigServerLoad(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// validate config has required fields
	if len(cfg.Security.WhitelistHost) == 0 || len(cfg.Security.WhitelistOrigin) == 0 {
		t.Fatal("config missing required whitelist host or origin")
	}

	server = fmt.Sprintf("http://%s:%s",
		cfg.Listener.BackendApi.Address, strconv.Itoa(int(cfg.Listener.BackendApi.Port)))

	if len(server) <= 0 {
		t.Fatal("server variable still empty")
	}

	// wait for server to be ready with proper health check (with required headers)
	t.Log("waiting for server to be ready...")
	maxRetries := 40
	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest(http.MethodGet, server+backend_api.BackendApiStatusHint, nil)
		if err != nil {
			time.Sleep(1000 * time.Millisecond)
			continue
		}

		// must include whitelist headers for status endpoint
		req.Host = cfg.Security.WhitelistHost[0]
		req.Header.Set("Origin", cfg.Security.WhitelistOrigin[0])

		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			t.Logf("server ready after %d retries", i+1)
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1000 * time.Millisecond)
	}

	t.Fatal("server failed to start within timeout - check server logs above for errors")
}

// stop backend server and cleanup resources properly
func teardownTestServer(t *testing.T) {
	t.Helper()

	if serverProcess != nil && serverProcess.Process != nil {
		t.Log("stopping server gracefully...")

		// try graceful shutdown first with sigterm
		if err := serverProcess.Process.Signal(syscall.SIGTERM); err != nil {
			t.Logf("failed to send sigterm: %v, forcing kill", err)
			// fallback to sigkill if sigterm fails
			if err := serverProcess.Process.Kill(); err != nil {
				t.Logf("failed to kill process: %v", err)
			}
		}

		// wait for process to exit with timeout
		done := make(chan error, 1)
		go func() {
			done <- serverProcess.Wait()
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Logf("server process exited with: %v", err)
			} else {
				t.Log("server process exited cleanly")
			}
		case <-time.After(5 * time.Second):
			t.Log("server did not exit within 5 seconds, forcing kill")
			if err := serverProcess.Process.Kill(); err != nil {
				t.Logf("failed to force kill: %v", err)
			}
			<-done // wait for wait() to return
		}
	}

	// cleanup binary file from project root
	binaryPath := filepath.Join(projectRoot, testBinary)
	if err := os.Remove(binaryPath); err != nil {
		t.Logf("warning: failed to remove test binary %s: %v", binaryPath, err)
	}

	// verify port is actually released before test completes
	// use bind approach instead of dial to avoid false positives
	port := cfg.Listener.BackendApi.Port
	address := fmt.Sprintf("%s:%d", cfg.Listener.BackendApi.Address, port)
	t.Logf("verifying port %s is released...", address)

	maxRetries := 50
	for i := 0; i < maxRetries; i++ {
		// try to bind to the port - if successful, it means the port is truly free
		listener, err := net.Listen("tcp", address)
		if err == nil {
			// port is free! close our test listener
			listener.Close()
			t.Logf("port %s successfully released after %d checks", address, i+1)
			return
		}

		// port is still in use, wait and retry
		time.Sleep(1000 * time.Millisecond)
	}

	// final check: see what's actually using the port
	t.Logf("port %s still appears to be in use, checking what's holding it...", address)
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Logf("processes using port %d:\n%s", port, string(output))
	} else {
		t.Logf("could not check port usage: %v", err)
	}

	t.Logf("warning: port %s may still be in time_wait state or held by another process", address)
}

// --------------------------------------------------------- //

// generate random test user data with valid email format
func generateTestData() {
	// use simple valid email format to avoid validation errors
	timestamp := time.Now().UnixNano()
	email = fmt.Sprintf("test%d@example.com", timestamp)

	// generate secure password
	genPassword, _ := pkg.GenRandomNumber(8, 12)
	password, _ = pkg.GenRandomAlphanumeric(genPassword)
}

// --------------------------------------------------------- //

func TestBackendApi_IntegrationSuite(t *testing.T) {
	// run all tests sequentially in one suite to maintain state dependency
	setupTestServer(t)
	defer teardownTestServer(t)

	t.Run("1_check_status", func(t *testing.T) {
		generateTestData()

		url := server + backend_api.BackendApiStatusHint

		if len(cfg.Security.WhitelistHost) <= 0 || len(cfg.Security.WhitelistOrigin) <= 0 {
			t.Fatal("whitelist host or whitelist origin can't be empty")
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			t.Fatalf("fail to make new request: %v", err)
		}

		req.Host = cfg.Security.WhitelistHost[0]
		req.Header.Set("Origin", cfg.Security.WhitelistOrigin[0])

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("can't do client request: %v", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expecting 200 but got %d; resp body: %s", resp.StatusCode, string(respBody))
		}
		t.Log("status check passed")
	})

	t.Run("2_create_account_user", func(t *testing.T) {
		url := server + backend_api_account.BackendApiAccountUserHint
		body := map[string]any{
			"email":    email,
			"password": password,
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatal("fail to make json marshal")
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			t.Fatalf("fail to make new request: %v", err)
		}

		req.Host = cfg.Security.WhitelistHost[0]
		req.Header.Set("Origin", cfg.Security.WhitelistOrigin[0])

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("can't do client request: %v", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expecting 200 but got %d; resp body: %s", resp.StatusCode, string(respBody))
		}
		t.Log("account creation passed")
	})

	t.Run("3_get_account_user_id", func(t *testing.T) {
		url := server + backend_api_account.BackendApiAccountUserHint + "?email=" + email

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			t.Fatalf("fail to make new request: %v", err)
		}

		req.Host = cfg.Security.WhitelistHost[0]
		req.Header.Set("Origin", cfg.Security.WhitelistOrigin[0])

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("can't do client request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expecting 200 but got %d", resp.StatusCode)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("fail to read resp body: %v; body: %s", err, string(respBody))
		}

		var raw map[string]any
		err = json.Unmarshal(respBody, &raw)
		if err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}

		okValue, okExists := raw["ok"]
		if !okExists {
			t.Fatal("ok field doesn't exists")
		}

		ok := okValue.(bool)
		if !ok {
			t.Fatal("ok field value must be true")
		}

		dataValue, dataExists := raw["data"]
		if !dataExists {
			t.Fatal("data field doesn't exists")
		}

		dataMap, dataMapOk := dataValue.(map[string]any)
		if !dataMapOk {
			t.Fatalf("data map must be an object; got %T = %v", dataMap, dataMap)
		}

		idValue, idExists := dataMap["id"]
		if !idExists {
			t.Fatal("expecting id field in data object")
		}

		idStr := strings.TrimSpace(idValue.(string))

		idUuid, err := uuid.Parse(idStr)
		if err != nil {
			t.Fatalf("fail to parse id of uuid: %v", err)
		}

		userId = idUuid
		authorizationData = base64.StdEncoding.EncodeToString([]byte(userId.String()))
		t.Log("get account user id passed")
	})

	t.Run("4_create_session", func(t *testing.T) {
		url := server + backend_api_auth.BackendApiAuthSessionHint

		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			t.Fatalf("fail to make new request: %v", err)
		}

		req.Host = cfg.Security.WhitelistHost[0]
		req.Header.Set("Origin", cfg.Security.WhitelistOrigin[0])
		req.Header.Set("Authorization",
			mw.AuthorizationHeadKey_bearer+" "+authorizationData)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("can't do client request: %v", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expecting 200 but got %d; resp body: %s", resp.StatusCode, string(respBody))
		}
		t.Log("session creation passed")
	})

	t.Run("5_update_email", func(t *testing.T) {
		timestamp := time.Now().UnixNano()
		newEmail := fmt.Sprintf("test-updated%d@example.com", timestamp)

		url := server + backend_api_account.BackendApiAccountUserHint
		body := map[string]any{
			"id":    userId.String(),
			"email": newEmail,
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatal("fail to make json marshal")
		}

		req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			t.Fatalf("fail to make new request: %v", err)
		}

		req.Host = cfg.Security.WhitelistHost[0]
		req.Header.Set("Origin", cfg.Security.WhitelistOrigin[0])
		req.Header.Set("Authorization",
			mw.AuthorizationHeadKey_bearer+" "+authorizationData)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("can't do client request: %v", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expecting 200 but got %d; resp body: %s", resp.StatusCode, string(respBody))
		}
		t.Log("email update passed")
	})

	t.Run("6_create_stash", func(t *testing.T) {
		url := server + backend_api_game1.BackendApiGame1StashHint
		stashGen, err := pkg.GenRandomAlphanumeric(6)
		if err != nil {
			t.Fatalf("fail to generate stash name: %v", err)
		}
		stashName = "Stash " + stashGen
		body := map[string]any{
			"name": stashName,
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatal("fail to make json marshal")
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			t.Fatalf("fail to make new request: %v", err)
		}

		req.Host = cfg.Security.WhitelistHost[0]
		req.Header.Set("Origin", cfg.Security.WhitelistOrigin[0])
		req.Header.Set("Authorization",
			mw.AuthorizationHeadKey_bearer+" "+authorizationData)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("can't do client request: %v", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expecting 200 but got %d; resp body: %s", resp.StatusCode, string(respBody))
		}
		t.Log("stash creation passed")
	})

	t.Run("7_update_stash", func(t *testing.T) {
		url := server + backend_api_game1.BackendApiGame1StashHint
		operand, err := pkg.GenRandomNumber(1, 2)
		if err != nil {
			t.Fatalf("fail to generate random number for operand: %v", err)
		}
		itemGenLength, err := pkg.GenRandomNumber(3, 9)
		if err != nil {
			t.Fatalf("fail to generate random number for itemGenLength: %v", err)
		}
		item, err := pkg.GenRandomAlphanumeric(itemGenLength)
		if err != nil {
			t.Fatalf("fail to generate item name: %v", err)
		}
		quantity, err := pkg.GenRandomNumber(1, 3)
		if err != nil {
			t.Fatalf("fail to generate random number for quantity: %v", err)
		}
		body := map[string]any{
			"name":     stashName,
			"operand":  operand,
			"item":     item,
			"quantity": quantity,
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatal("fail to make json marshal")
		}

		req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			t.Fatalf("fail to make new request: %v", err)
		}

		req.Host = cfg.Security.WhitelistHost[0]
		req.Header.Set("Origin", cfg.Security.WhitelistOrigin[0])
		req.Header.Set("Authorization",
			mw.AuthorizationHeadKey_bearer+" "+authorizationData)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("can't do client request: %v", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expecting 200 got %d; resp body: %s", resp.StatusCode, string(respBody))
		}
		t.Log("stash update passed")
	})
}
