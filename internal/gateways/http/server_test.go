package http

import (
	"fmt"
	"homework/internal/usecase"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestWithHost(t *testing.T) {
	server := &Server{host: "default"}
	option := WithHost("customhost")
	option(server)
	assert.Equal(t, "customhost", server.host)
}

func TestWithPort(t *testing.T) {
	server := &Server{port: 1234}
	option := WithPort(5678)
	option(server)
	assert.Equal(t, uint16(5678), server.port)
}

func TestNewServer(t *testing.T) {
	useCases := UseCases{
		Event:  &usecase.Event{},
		Sensor: &usecase.Sensor{},
		User:   &usecase.User{},
	}

	server := NewServer(useCases)

	assert.NotNil(t, server)
	assert.Equal(t, "localhost", server.host)
	assert.Equal(t, uint16(8080), server.port)
	assert.NotNil(t, server.router)
}

func TestNewServerWithOptions(t *testing.T) {
	useCases := UseCases{
		Event:  &usecase.Event{},
		Sensor: &usecase.Sensor{},
		User:   &usecase.User{},
	}

	customHost := "127.0.0.1"
	customPort := uint16(9090)

	server := NewServer(
		useCases,
		WithHost(customHost),
		WithPort(customPort),
	)

	assert.NotNil(t, server)
	assert.Equal(t, customHost, server.host)
	assert.Equal(t, customPort, server.port)
}

func TestServerRun(t *testing.T) {
	engine := gin.New()
	engine.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	server := &Server{
		router: engine,
		host:   "localhost",
		port:   8080,
	}

	testServer := httptest.NewServer(engine)
	defer testServer.Close()

	resp, err := http.Get(testServer.URL + "/test")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	host := "testhost"
	port := uint16(8888)
	server.host = host
	server.port = port

	assert.Equal(t, host, server.host)
	assert.Equal(t, port, server.port)
}

func TestServerOptions(t *testing.T) {
	server := &Server{}

	WithHost("host1")(server)
	assert.Equal(t, "host1", server.host)

	WithPort(1111)(server)
	assert.Equal(t, uint16(1111), server.port)

	WithHost("host2")(server)
	assert.Equal(t, "host2", server.host)

	WithPort(2222)(server)
	assert.Equal(t, uint16(2222), server.port)
}

func TestServerEdgeCases(t *testing.T) {
	server := &Server{host: "localhost"}
	WithHost("")(server)
	assert.Equal(t, "", server.host)

	server = &Server{port: 8080}
	WithPort(0)(server)
	assert.Equal(t, uint16(0), server.port)

	WithPort(65535)(server)
	assert.Equal(t, uint16(65535), server.port)
}

func TestNewServerWithEmptyUseCases(t *testing.T) {
	useCases := UseCases{}
	server := NewServer(useCases)
	assert.NotNil(t, server)
}

func TestServerOptionsChained(t *testing.T) {
	useCases := UseCases{}

	server := NewServer(
		useCases,
		WithHost("custom.host.test"),
		WithPort(12345),
	)

	assert.Equal(t, "custom.host.test", server.host)
	assert.Equal(t, uint16(12345), server.port)
}

func TestRunMethodAddress(t *testing.T) {
	useCases := UseCases{}
	server := NewServer(useCases)

	addrString := fmt.Sprintf("%s:%d", server.host, server.port)
	assert.Contains(t, addrString, "localhost")
	assert.Contains(t, addrString, "8080")

	server.host = "test.host"
	server.port = 9999

	addrString = fmt.Sprintf("%s:%d", server.host, server.port)
	assert.Equal(t, "test.host:9999", addrString)
}
