package service

import "context"

type ATGServer struct {
}

func (s *ATGServer) CreateSession(context.Context, *SessionInfo) (*Success, error) {
	return nil, nil
}
func (s *ATGServer) GetGameServerInfo(context.Context, *SessionKey) (*ServerInfo, error) {
	return nil, nil
}
func (s *ATGServer) GetRoomList(context.Context, *SessionKey) (*RoomList, error) {
	return nil, nil
}

type ATGClient struct {
}

func (c *ATGClient) CreateSession(ctx context.Context, in *SessionInfo, opts ...grpc.CallOption) (*Success, error) {
	return nil, nil
}
func (c *ATGClient) GetGameServerInfo(ctx context.Context, in *SessionKey, opts ...grpc.CallOption) (*ServerInfo, error) {
	return nil, nil
}
func (c *ATGClient) GetRoomList(ctx context.Context, in *SessionKey, opts ...grpc.CallOption) (*RoomList, error) {
	return nil, nil
}
