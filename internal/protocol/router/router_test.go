package router

import (
	"context"
	"fmt"
	"testing"

	"github.com/meta-mcp/meta-mcp-server/internal/protocol/jsonrpc"
)

// Mock handler for testing
type mockHandler struct {
	method string
	result interface{}
}

func (m *mockHandler) Handle(ctx context.Context, request *jsonrpc.Request) *jsonrpc.Response {
	return jsonrpc.NewResponse(m.result, request.ID)
}

// Mock notification handler for testing
type mockNotificationHandler struct {
	called bool
	method string
}

func (m *mockNotificationHandler) HandleNotification(ctx context.Context, notification *jsonrpc.Notification) {
	m.called = true
	m.method = notification.Method
}

func TestRouter_Register(t *testing.T) {
	router := New()
	handler := &mockHandler{method: "test", result: "success"}

	router.Register("test", handler)

	if !router.HasMethod("test") {
		t.Error("Expected method 'test' to be registered")
	}

	methods := router.GetRegisteredMethods()
	if len(methods) != 1 || methods[0] != "test" {
		t.Errorf("Expected registered methods to be ['test'], got %v", methods)
	}
}

func TestRouter_RegisterFunc(t *testing.T) {
	router := New()

	router.RegisterFunc("test", func(ctx context.Context, request *jsonrpc.Request) *jsonrpc.Response {
		return jsonrpc.NewResponse("function result", request.ID)
	})

	if !router.HasMethod("test") {
		t.Error("Expected method 'test' to be registered")
	}
}

func TestRouter_Handle(t *testing.T) {
	router := New()
	handler := &mockHandler{method: "test", result: "success"}
	router.Register("test", handler)

	request := jsonrpc.NewRequest("test", nil, "req-1")
	response := router.Handle(context.Background(), request)

	if response.ID != "req-1" {
		t.Errorf("Expected response ID 'req-1', got %v", response.ID)
	}

	if response.Result != "success" {
		t.Errorf("Expected result 'success', got %v", response.Result)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}
}

func TestRouter_Handle_UnknownMethod(t *testing.T) {
	router := New()

	request := jsonrpc.NewRequest("unknown", nil, "req-1")
	response := router.Handle(context.Background(), request)

	if response.ID != "req-1" {
		t.Errorf("Expected response ID 'req-1', got %v", response.ID)
	}

	if response.Error == nil {
		t.Error("Expected error for unknown method")
	}

	if response.Error.Code != jsonrpc.ErrorCodeMethodNotFound {
		t.Errorf("Expected error code %d, got %d", jsonrpc.ErrorCodeMethodNotFound, response.Error.Code)
	}
}

func TestRouter_Handle_DefaultHandler(t *testing.T) {
	router := New()
	defaultHandler := &mockHandler{method: "default", result: "default result"}
	router.SetDefaultHandler(defaultHandler)

	request := jsonrpc.NewRequest("unknown", nil, "req-1")
	response := router.Handle(context.Background(), request)

	if response.Result != "default result" {
		t.Errorf("Expected result 'default result', got %v", response.Result)
	}

	if response.Error != nil {
		t.Errorf("Expected no error, got %v", response.Error)
	}
}

func TestRouter_RegisterNotification(t *testing.T) {
	router := New()
	handler := &mockNotificationHandler{}

	router.RegisterNotification("notify", handler)

	if !router.HasNotificationMethod("notify") {
		t.Error("Expected notification method 'notify' to be registered")
	}

	methods := router.GetRegisteredNotificationMethods()
	if len(methods) != 1 || methods[0] != "notify" {
		t.Errorf("Expected registered notification methods to be ['notify'], got %v", methods)
	}
}

func TestRouter_HandleNotification(t *testing.T) {
	router := New()
	handler := &mockNotificationHandler{}
	router.RegisterNotification("notify", handler)

	notification := jsonrpc.NewNotification("notify", nil)
	router.HandleNotification(context.Background(), notification)

	if !handler.called {
		t.Error("Expected notification handler to be called")
	}

	if handler.method != "notify" {
		t.Errorf("Expected handler method 'notify', got %s", handler.method)
	}
}

func TestRouter_HandleNotification_UnknownMethod(t *testing.T) {
	router := New()

	// This should not panic or cause errors
	notification := jsonrpc.NewNotification("unknown", nil)
	router.HandleNotification(context.Background(), notification)
}

func TestRouter_HandleNotification_DefaultHandler(t *testing.T) {
	router := New()
	defaultHandler := &mockNotificationHandler{}
	router.SetDefaultNotificationHandler(defaultHandler)

	notification := jsonrpc.NewNotification("unknown", nil)
	router.HandleNotification(context.Background(), notification)

	if !defaultHandler.called {
		t.Error("Expected default notification handler to be called")
	}
}

func TestRouter_Unregister(t *testing.T) {
	router := New()
	handler := &mockHandler{method: "test", result: "success"}
	router.Register("test", handler)

	if !router.HasMethod("test") {
		t.Error("Expected method 'test' to be registered")
	}

	router.Unregister("test")

	if router.HasMethod("test") {
		t.Error("Expected method 'test' to be unregistered")
	}
}

func TestRouter_UnregisterNotification(t *testing.T) {
	router := New()
	handler := &mockNotificationHandler{}
	router.RegisterNotification("notify", handler)

	if !router.HasNotificationMethod("notify") {
		t.Error("Expected notification method 'notify' to be registered")
	}

	router.UnregisterNotification("notify")

	if router.HasNotificationMethod("notify") {
		t.Error("Expected notification method 'notify' to be unregistered")
	}
}

func TestRouter_Clear(t *testing.T) {
	router := New()
	handler := &mockHandler{method: "test", result: "success"}
	notificationHandler := &mockNotificationHandler{}

	router.Register("test", handler)
	router.RegisterNotification("notify", notificationHandler)
	router.SetDefaultHandler(handler)
	router.SetDefaultNotificationHandler(notificationHandler)

	router.Clear()

	if router.HasMethod("test") {
		t.Error("Expected all methods to be cleared")
	}

	if router.HasNotificationMethod("notify") {
		t.Error("Expected all notification methods to be cleared")
	}

	stats := router.GetStats()
	if stats.HasDefaultHandler {
		t.Error("Expected default handler to be cleared")
	}

	if stats.HasDefaultNotificationHandler {
		t.Error("Expected default notification handler to be cleared")
	}
}

func TestRouter_GetStats(t *testing.T) {
	router := New()
	handler := &mockHandler{method: "test", result: "success"}
	notificationHandler := &mockNotificationHandler{}

	router.Register("test1", handler)
	router.Register("test2", handler)
	router.RegisterNotification("notify1", notificationHandler)
	router.SetDefaultHandler(handler)

	stats := router.GetStats()

	if stats.RegisteredMethods != 2 {
		t.Errorf("Expected 2 registered methods, got %d", stats.RegisteredMethods)
	}

	if stats.RegisteredNotificationMethods != 1 {
		t.Errorf("Expected 1 registered notification method, got %d", stats.RegisteredNotificationMethods)
	}

	if !stats.HasDefaultHandler {
		t.Error("Expected default handler to be set")
	}

	if stats.HasDefaultNotificationHandler {
		t.Error("Expected default notification handler to not be set")
	}
}

func TestRouter_ThreadSafety(t *testing.T) {
	router := New()
	handler := &mockHandler{method: "test", result: "success"}

	// Test concurrent registration and handling
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			router.Register("test", handler)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			request := jsonrpc.NewRequest("test", nil, i)
			router.Handle(context.Background(), request)
		}
		done <- true
	}()

	<-done
	<-done

	// If we get here without panicking, the test passes
}

func TestHandlerFunc(t *testing.T) {
	handlerFunc := HandlerFunc(func(ctx context.Context, request *jsonrpc.Request) *jsonrpc.Response {
		return jsonrpc.NewResponse("function result", request.ID)
	})

	request := jsonrpc.NewRequest("test", nil, "req-1")
	response := handlerFunc.Handle(context.Background(), request)

	if response.Result != "function result" {
		t.Errorf("Expected result 'function result', got %v", response.Result)
	}
}

func TestNotificationHandlerFunc(t *testing.T) {
	called := false
	handlerFunc := NotificationHandlerFunc(func(ctx context.Context, notification *jsonrpc.Notification) {
		called = true
	})

	notification := jsonrpc.NewNotification("test", nil)
	handlerFunc.HandleNotification(context.Background(), notification)

	if !called {
		t.Error("Expected notification handler function to be called")
	}
}

// Benchmarks for router performance
func BenchmarkRouterHandle(b *testing.B) {
	router := New()
	router.RegisterFunc("test.method", func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
		return jsonrpc.NewResponse(map[string]interface{}{"success": true}, req.ID)
	})

	request := jsonrpc.NewRequest("test.method", map[string]interface{}{"key": "value"}, 1)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		response := router.Handle(ctx, request)
		if response.Error != nil {
			b.Fatal("Unexpected error in benchmark")
		}
	}
}

func BenchmarkRouterHandleNotFound(b *testing.B) {
	router := New()
	request := jsonrpc.NewRequest("nonexistent.method", nil, 1)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		response := router.Handle(ctx, request)
		if response.Error == nil {
			b.Fatal("Expected method not found error")
		}
	}
}

func BenchmarkRouterHandleNotification(b *testing.B) {
	router := New()
	called := 0
	router.RegisterNotificationFunc("test.notification", func(ctx context.Context, notif *jsonrpc.Notification) {
		called++
	})

	notification := jsonrpc.NewNotification("test.notification", map[string]interface{}{"key": "value"})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.HandleNotification(ctx, notification)
	}
}

func BenchmarkRouterConcurrentAccess(b *testing.B) {
	router := New()

	// Register multiple handlers
	for i := 0; i < 10; i++ {
		method := fmt.Sprintf("test.method%d", i)
		router.RegisterFunc(method, func(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
			return jsonrpc.NewResponse(map[string]interface{}{"method": req.Method}, req.ID)
		})
	}

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			method := fmt.Sprintf("test.method%d", i%10)
			request := jsonrpc.NewRequest(method, nil, i)
			response := router.Handle(ctx, request)
			if response.Error != nil {
				b.Fatal("Unexpected error in concurrent benchmark")
			}
			i++
		}
	})
}
