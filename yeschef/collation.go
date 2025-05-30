package yeschef

import (
	"log"
	"sync"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/reugn/go-quartz/quartz"
)

type ChefsKiss struct {
	mu             sync.RWMutex
	apps           map[int64]*CmdServer // Keyed by user_id
	mcpApps        map[int64]*McpServer // Keyed by app_id
	RunningMan     *RunningMan
	NowQueue       *jobQueue
	NowScheduler   *quartz.StdScheduler
	InQueue        *jobQueue
	InScheduler    *quartz.StdScheduler
	EveryQueue     *jobQueue
	EveryScheduler *quartz.StdScheduler
}

func (x *ChefsKiss) ReadInstance(user_id int64) *CmdServer {
	x.mu.RLock()
	defer x.mu.RUnlock()
	server, ok := x.apps[user_id]
	if ok {
		return server
	}
	return nil
}

func (x *ChefsKiss) DeleteInstance(user_id int64) {
	x.mu.Lock()
	defer x.mu.Unlock()
	_, ok := x.apps[user_id]
	if ok {
		// TODO: Should we also stop the server's Run loop?
		delete(x.apps, user_id)
	}
}

// CreateInstance ensures a CmdServer exists for the given user ID.
// If an instance already exists it is returned, otherwise a new one is
// created, started and stored before being returned.
func (x *ChefsKiss) CreateInstance(userID int64) *CmdServer {
	x.mu.Lock()
	defer x.mu.Unlock()

	if server, ok := x.apps[userID]; ok {
		return server
	}

	srv := NewServer()
	go srv.Run()
	x.apps[userID] = srv
	return srv
}

// CreateMcpAppInstance ensures an MCP server exists for the given app ID.
func (x *ChefsKiss) CreateMcpAppInstance(appID int64) *McpServer {
	x.mu.Lock()
	defer x.mu.Unlock()

	if srv, ok := x.mcpApps[appID]; ok {
		return srv
	}

	app, err := models.AppByID(appID)
	if err != nil {
		log.Printf("CreateMcpAppInstance AppByID: %v", err)
		return nil
	}

	srv := NewMcpServer(appID, app.UUID, app.YAMLShared)
	go srv.Run()
	x.mcpApps[appID] = srv
	return srv
}

// ReadMcpAppInstance returns the MCP server for the given app ID.
func (x *ChefsKiss) ReadMcpAppInstance(appID int64) *McpServer {
	x.mu.RLock()
	defer x.mu.RUnlock()
	return x.mcpApps[appID]
}

// Ensure maps are initialized (this should happen where ChefsKiss is instantiated)
// Example initialization (actual location might vary):
// func NewChefsKiss() *ChefsKiss {
//     return &ChefsKiss{
//         apps:           make(map[int64]*CmdServer),
//         // ... other fields ...
//     }
// }
