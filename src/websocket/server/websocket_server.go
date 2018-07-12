package server

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Tanibox/tania-core/config"
	"github.com/Tanibox/tania-core/proto"
	"github.com/Tanibox/tania-core/src/assets/domain/service"
	"github.com/Tanibox/tania-core/src/assets/query"
	queryMysql "github.com/Tanibox/tania-core/src/assets/query/mysql"
	querySqlite "github.com/Tanibox/tania-core/src/assets/query/sqlite"
	"github.com/Tanibox/tania-core/src/assets/repository"
	repoMysql "github.com/Tanibox/tania-core/src/assets/repository/mysql"
	repoSqlite "github.com/Tanibox/tania-core/src/assets/repository/sqlite"
	"github.com/Tanibox/tania-core/src/eventbus"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
)

// WebsocketServer ties the routes and handlers with injected dependencies
type WebsocketServer struct {
	DeviceEventRepo  repository.DeviceEventRepository
	DeviceEventQuery query.DeviceEventQuery
	DeviceReadRepo   repository.DeviceReadRepository
	DeviceReadQuery  query.DeviceReadQuery
	DeviceService    service.DeviceService

	EventBus eventbus.TaniaEventBus
}

// NewWebsocketServer initializes WebsocketServer's dependencies and create new WebsocketServer struct
func NewWebsocketServer(
	db *sql.DB,
	eventBus eventbus.TaniaEventBus,
) (*WebsocketServer, error) {
	userServer := &WebsocketServer{
		EventBus: eventBus,
	}

	switch *config.Config.TaniaPersistenceEngine {
	case config.DB_SQLITE:
		userServer.DeviceEventRepo = repoSqlite.NewDeviceEventRepositorySqlite(db)
		userServer.DeviceEventQuery = querySqlite.NewDeviceEventQuerySqlite(db)
		userServer.DeviceReadRepo = repoSqlite.NewDeviceReadRepositorySqlite(db)
		userServer.DeviceReadQuery = querySqlite.NewDeviceReadQuerySqlite(db)

		userServer.DeviceService = service.DeviceService{
			DeviceReadQuery: userServer.DeviceReadQuery,
		}

	case config.DB_MYSQL:
		userServer.DeviceEventRepo = repoMysql.NewDeviceEventRepositoryMysql(db)
		userServer.DeviceEventQuery = queryMysql.NewDeviceEventQueryMysql(db)
		userServer.DeviceReadRepo = repoMysql.NewDeviceReadRepositoryMysql(db)
		userServer.DeviceReadQuery = queryMysql.NewDeviceReadQueryMysql(db)

		userServer.DeviceService = service.DeviceService{
			DeviceReadQuery: userServer.DeviceReadQuery,
		}

	}

	userServer.InitSubscriber()

	return userServer, nil
}

// InitSubscriber defines the mapping of which event this domain listen with their handler
func (s *WebsocketServer) InitSubscriber() {

}

// Mount defines the WebsocketServer's endpoints with its handlers
func (s *WebsocketServer) Mount(g *echo.Group) {
	g.GET("/devices/sensor", s.WebsocketReceiveDeviceData)
}

func (s *WebsocketServer) WebsocketReceiveDeviceData(c echo.Context) error {
	var upgrader = websocket.Upgrader{}
	ws, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			c.Logger().Error(err)
		}

		m := &proto.SensorData{}
		if err := protobuf.Unmarshal(msg, m); err != nil {
			log.Fatalln("Failed to parse data:", err)
		}

		fmt.Printf("websocket: %s\n", m)
	}
}
