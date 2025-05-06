package http

import (
	"context"
	"encoding/json"
	"homework/internal/domain"
	"homework/internal/usecase"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/coder/websocket"
)

type testSuite struct {
	suite.Suite

	ctrl *gomock.Controller
}

func (t *testSuite) SetupSuite() {
	t.ctrl = gomock.NewController(t.T())
}

func (t *testSuite) TestWebSocketConnection() {
	engine := gin.Default()

	erMock := usecase.NewMockEventRepository(t.ctrl)
	erMock.EXPECT().GetLastEventBySensorID(gomock.Any(), gomock.Eq(int64(1))).Return(&domain.Event{SensorID: 1, Payload: 100}, nil).Times(1)
	srMock := usecase.NewMockSensorRepository(t.ctrl)
	srMock.EXPECT().GetSensorByID(gomock.Any(), gomock.Eq(int64(1))).Return(&domain.Sensor{ID: 1}, nil).Times(1)
	urMock := usecase.NewMockUserRepository(t.ctrl)
	sorMock := usecase.NewMockSensorOwnerRepository(t.ctrl)

	uc := UseCases{
		Event:  usecase.NewEvent(erMock, srMock),
		Sensor: usecase.NewSensor(srMock),
		User:   usecase.NewUser(urMock, sorMock, srMock),
	}

	ws := NewWebSocketHandler(uc)
	setupRouter(engine, uc, ws)

	srv := httptest.NewServer(engine)
	defer srv.Close()

	srvURL, _ := url.Parse(srv.URL)
	srvURL.Scheme = "ws"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, srvURL.String()+"/sensors/1/events", nil)
	require.NoError(t.T(), err)
	op, msg, err := conn.Read(ctx)
	require.NoError(t.T(), err)
	require.Equal(t.T(), websocket.MessageText, op)
	var event domain.Event
	require.NoError(t.T(), json.Unmarshal(msg, &event))

	require.Equal(t.T(), int64(1), event.SensorID)
	require.Equal(t.T(), int64(100), event.Payload)
}

func (t *testSuite) TestWebSocketConnectionFail() {
	engine := gin.Default()

	erMock := usecase.NewMockEventRepository(t.ctrl)
	srMock := usecase.NewMockSensorRepository(t.ctrl)
	srMock.EXPECT().GetSensorByID(gomock.Any(), gomock.Eq(int64(1))).Return(nil, usecase.ErrSensorNotFound).Times(1)
	urMock := usecase.NewMockUserRepository(t.ctrl)
	sorMock := usecase.NewMockSensorOwnerRepository(t.ctrl)

	uc := UseCases{
		Event:  usecase.NewEvent(erMock, srMock),
		Sensor: usecase.NewSensor(srMock),
		User:   usecase.NewUser(urMock, sorMock, srMock),
	}

	ws := NewWebSocketHandler(uc)
	setupRouter(engine, uc, ws)

	srv := httptest.NewServer(engine)
	defer srv.Close()

	srvURL, _ := url.Parse(srv.URL)
	srvURL.Scheme = "ws"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, srvURL.String()+"/sensors/1/events", nil)
	require.Error(t.T(), err)
	assert.Equal(t.T(), http.StatusNotFound, resp.StatusCode)
}

func (t *testSuite) TestWebSocketShutdown_Server() {
	engine := gin.Default()
	erMock := usecase.NewMockEventRepository(t.ctrl)
	srMock := usecase.NewMockSensorRepository(t.ctrl)
	srMock.EXPECT().GetSensorByID(gomock.Any(), gomock.Eq(int64(2))).Return(&domain.Sensor{ID: 2}, nil).Times(1)
	urMock := usecase.NewMockUserRepository(t.ctrl)
	sorMock := usecase.NewMockSensorOwnerRepository(t.ctrl)

	uc := UseCases{
		Event:  usecase.NewEvent(erMock, srMock),
		Sensor: usecase.NewSensor(srMock),
		User:   usecase.NewUser(urMock, sorMock, srMock),
	}

	ws := NewWebSocketHandler(uc)
	setupRouter(engine, uc, ws)

	srv := httptest.NewServer(engine)
	defer srv.Close()

	srvURL, _ := url.Parse(srv.URL)
	srvURL.Scheme = "ws"
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, srvURL.String()+"/sensors/2/events", nil)
	require.NoError(t.T(), err)
	go func() {
		time.Sleep(time.Millisecond * 100)
		assert.NoError(t.T(), ws.Shutdown())
	}()
	op, _, err := conn.Read(ctx)
	assert.Equal(t.T(), websocket.MessageType(0), op)
	assert.Error(t.T(), websocket.CloseError{Code: websocket.StatusNormalClosure, Reason: "server shutting down"}, err)
}

func (t *testSuite) TestWebSocketShutdown_Client() {
	engine := gin.Default()
	erMock := usecase.NewMockEventRepository(t.ctrl)
	srMock := usecase.NewMockSensorRepository(t.ctrl)
	srMock.EXPECT().GetSensorByID(gomock.Any(), gomock.Eq(int64(2))).Return(&domain.Sensor{ID: 2}, nil).Times(1)
	urMock := usecase.NewMockUserRepository(t.ctrl)
	sorMock := usecase.NewMockSensorOwnerRepository(t.ctrl)

	uc := UseCases{
		Event:  usecase.NewEvent(erMock, srMock),
		Sensor: usecase.NewSensor(srMock),
		User:   usecase.NewUser(urMock, sorMock, srMock),
	}

	ws := NewWebSocketHandler(uc)
	setupRouter(engine, uc, ws)

	srv := httptest.NewServer(engine)
	defer srv.Close()

	srvURL, _ := url.Parse(srv.URL)
	srvURL.Scheme = "ws"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, srvURL.String()+"/sensors/2/events", nil)
	require.NoError(t.T(), err)
	go func() {
		time.Sleep(time.Millisecond * 100)
		assert.NoError(t.T(), conn.Close(websocket.StatusNormalClosure, "bye-bye"))
	}()
	op, _, err := conn.Read(ctx)
	assert.Equal(t.T(), websocket.MessageType(0), op)
	assert.Error(t.T(), websocket.CloseError{Code: websocket.StatusNormalClosure, Reason: "bye-bye"}, err)
	assert.NoError(t.T(), ws.Shutdown())
}

func TestWebSocketHandler(t *testing.T) {
	ts := new(testSuite)
	defer func() {
		ts.ctrl.Finish()
	}()

	suite.Run(t, ts)
}
