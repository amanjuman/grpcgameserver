package service

import "context"

type CTGServer struct {
}

func (s *CTGServer) EnterRoom(context.Context, *ServerInfo) (*Success, error) {
	return nil, nil
}

func (s *CTGServer) LeaveRoom(context.Context, *Empty) (*Success, error) {
	return nil, nil
}

func (s *CTGServer) PlayerInput(ClientToGame_PlayerInputServer) error {
	return nil
}

func (s *CTGServer) UpdateRoomPrepareView(*Empty, ClientToGame_UpdateRoomPrepareViewServer) error {
	return nil
}
func (s *CTGServer) UpdateGameFrame(*Empty, ClientToGame_UpdateGameFrameServer) error {
	return nil
}
