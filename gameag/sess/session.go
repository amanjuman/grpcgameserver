package sess

import (
	//"github.com/amanjuman/grpcgameserver/entity"
	. "github.com/amanjuman/grpcgameserver/message"
	"github.com/amanjuman/grpcgameserver/user"
	. "github.com/amanjuman/grpcgameserver/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"strconv"
	"sync"
)

type MsgChannel struct {
	DataCh     chan (interface{})
	StopSignal chan (struct{})
}

func (m *MsgChannel) Close() {
	select {
	case <-m.StopSignal:
		return
	default:
		close(m.StopSignal)
	}
}

func NewMsgChannel(bufferNumber int32) *MsgChannel {
	return &MsgChannel{
		DataCh:     make(chan (interface{}), bufferNumber),
		StopSignal: make(chan (struct{}), 1),
	}
}

func NewMsgChannelManager() *MsgChannelManager {
	return &MsgChannelManager{
		make(map[string]*MsgChannel),
	}
}

type MsgChannelManager struct {
	c map[string]*MsgChannel
}

func (m *MsgChannelManager) AddMsgChan(name string, bufferNumber int32) bool {
	if _, ok := m.c[name]; ok {
		return false
	}
	m.c[name] = NewMsgChannel(bufferNumber)
	return true
}

func (m *MsgChannelManager) GetMsgChan(name string) *MsgChannel {
	return m.c[name]
}

func (m *MsgChannelManager) CloseMsgChan(name string) {
	if ch, ok := m.c[name]; ok {
		ch.Close()
		delete(m.c, name)
	}
}

type sessionManager struct {
	Sessions           map[int64]*Session
	UserNameMapSession map[string]int64
	sync.RWMutex
}

func (sm *sessionManager) CleanSession(id int64) {
	if s, ok := sm.Sessions[id]; ok {
		s.Room.LeaveRoom(s)
	}
}

func (sm *sessionManager) MakeSession() int64 {
	s := NewSession()
	sm.Lock()
	sm.Sessions[s.Info.Uuid] = s
	sm.Unlock()
	return s.Info.Uuid
}

func (sm *sessionManager) GetSession(md metadata.MD) *Session {
	mdid := md.Get("sess-id")
	if len(mdid) == 0 {
		return nil
	}
	id, err := strconv.ParseInt(mdid[0], 10, 64)
	if err != nil {
		return nil
	}
	s, ok := sm.Sessions[id]
	if !ok {
		return nil
	}
	s.RLock()
	if s.User != nil {
		uname := md.Get("uname")
		if len(uname) == 0 {
			s.RUnlock()
			return nil
		} else if s.User.UserInfo.UserName != uname[0] {
			s.RUnlock()
			return nil
		}
	} else {
		uname := md.Get("uname")
		if len(uname) != 0 {
			return nil
		}
	}
	s.RUnlock()
	return s
}
func NewSession() *Session {
	s := &Session{
		Info:              &SessionInfo{},
		MsgChannelManager: NewMsgChannelManager(),
		PlayerInfo:        &PlayerInfo{},
	}
	for i := int32(SessionInfo_NoSession); i <= int32(SessionInfo_GameServerWaitReconnect); i++ {
		ss := SessionStateFactory.makeSessionState(s, SessionInfo_SessionState(i))
		s.States = append(s.States, ss)
	}
	s.SetState(0)
	s.State.CreateSession()
	return s
}

type Session struct {
	Info       *SessionInfo
	State      SessionState
	SessionKey int64
	User       *user.User
	States     []SessionState
	Room       *Room
	*MsgChannelManager
	sync.RWMutex
	PlayerInfo *PlayerInfo
	TeamNo     int32
	IsReady    bool
	ServerInfo *ServerInfo
}

func (s *Session) GetPlayerInfo() *PlayerInfo {
	if s.User == nil {
		return nil
	}
	s.PlayerInfo.UserName = s.User.UserInfo.UserName
	s.PlayerInfo.UserId = s.User.UserInfo.Uuid
	if s.User.UserInfo.UsedCharacter == int64(0) {
		for id, c := range s.User.UserInfo.OwnCharacter {
			s.User.UserInfo.UsedCharacter = id
			s.PlayerInfo.Character = c
			break
		}
	} else {
		s.PlayerInfo.Character = s.User.UserInfo.OwnCharacter[s.User.UserInfo.UsedCharacter]
	}
	s.PlayerInfo.TeamNo = s.TeamNo
	s.PlayerInfo.IsReady = s.IsReady
	return s.PlayerInfo
}

func (s *Session) SetState(state_index int32) {
	s.State = s.States[state_index]
}

func (s *Session) GetSessionInfo() *SessionInfo {
	info := &UserInfo{}
	if s.User != nil {
		info = s.User.GetInfo()
	}
	return &SessionInfo{
		Uuid:     s.Info.Uuid,
		UserInfo: info,
		State:    s.State.GetStateCode(),
	}
}

func (s *Session) GetSessionCache() *SessionCache {
	return &SessionCache{
		GameServerInfo: s.ServerInfo,
		SessionInfo:    s.GetSessionInfo(),
	}
}

type SessionState interface {
	SetSession(s *Session) bool
	SetStateCode(SessionInfo_SessionState)
	GetStateCode() SessionInfo_SessionState
	CreateSession() int64
	Login(uname string, pswd string) *user.User
	Logout() bool
	Regist(uname string, pswd string, info ...string) bool
	CreateRoom(setting *RoomSetting) bool
	EnterRoom(roomId int64) bool
	DeleteRoom() bool
	ReadyRoom() bool
	LeaveRoom() bool
	StartRoom() bool
	SettingCharacter(*CharacterSetting) bool
	SettingRoom() bool
	CancelReady() bool
	EndRoom() bool
	String() string
	Lock()
	Unlock()
}

func (sb *SessionStateBase) SetSession(s *Session) bool {
	if sb.Session != nil {
		return false
	}
	sb.Session = s
	return true
}

func (sb *SessionStateBase) String() string {
	return SessionInfo_SessionState_name[int32(sb.StateCode)]
}

func (sb *SessionStateBase) SetStateCode(code SessionInfo_SessionState) {
	sb.StateCode = code
}
func (sb *SessionStateBase) GetStateCode() SessionInfo_SessionState {
	return sb.StateCode
}
func (sb *SessionStateBase) CreateSession() int64 {
	return 0
}
func (sb *SessionStateBase) Login(uname string, pswd string) *user.User {
	return nil
}
func (sb *SessionStateBase) Logout() bool {
	return false
}
func (sb *SessionStateBase) Regist(uname string, pswd string, info ...string) bool {
	return false
}
func (sb *SessionStateBase) CreateRoom(setting *RoomSetting) bool {
	return false
}
func (sb *SessionStateBase) EnterRoom(roomId int64) bool {
	return false
}
func (sb *SessionStateBase) DeleteRoom() bool {
	return false
}
func (sb *SessionStateBase) ReadyRoom() bool {
	return false
}
func (sb *SessionStateBase) LeaveRoom() bool {
	return false
}
func (sb *SessionStateBase) StartRoom() bool {
	return false
}
func (sb *SessionStateBase) SettingCharacter(*CharacterSetting) bool {
	return false
}
func (sb *SessionStateBase) SettingRoom() bool {
	return false
}
func (sb *SessionStateBase) EndRoom() bool {
	return false
}

func (sb *SessionStateBase) CancelReady() bool {
	return false
}

type SessionStateBase struct {
	StateCode SessionInfo_SessionState
	Session   *Session
	sync.RWMutex
}

type NoSessionState struct {
	SessionStateBase
}

func (ss *NoSessionState) CreateSession() int64 {
	//TODO
	s := ss.Session
	s.Lock()
	s.Info.Uuid, _ = Uid.NewId(SESSION_ID)
	uuid := s.Info.Uuid
	ss.Session.SetState(int32(ss.StateCode) + 1)
	s.Unlock()
	return uuid
}

type GuestSessionState struct {
	SessionStateBase
}

func (ss *GuestSessionState) Regist(uname string, pswd string, info ...string) bool {
	//TODO
	in := &RegistInput{UserName: uname, Pswd: pswd}
	_, err := user.Manager.Regist(in)
	if err != nil {
		return false
	}
	return true
}

func (ss *GuestSessionState) Login(uname string, pswd string) *user.User {
	//TODO

	in := &LoginInput{UserName: uname, Pswd: pswd}
	userInfo, err := user.Manager.Login(in)
	if err != nil {
		log.Warn(err)
	}
	if userInfo == nil {
		return nil
	}
	user := user.Manager.UserOnline[userInfo.Uuid]
	ss.Session.User = user
	ss.Session.SetState(int32(ss.StateCode) + 1)
	log.Info("user state:", ss.Session.State)
	ss.Session.AddMsgChan("RoomList", 10)
	RoomManager.AddIdleUserMsgChan(ss.Session.GetMsgChan("RoomList"))
	RoomManager.UpdateRoomList()
	if id, ok := Manager.UserNameMapSession[uname]; ok {
		Manager.CleanSession(id)
	}
	Manager.UserNameMapSession[uname] = ss.Session.Info.Uuid
	return user
}

type UserIdleSessionState struct {
	SessionStateBase
}

func (ss *UserIdleSessionState) CreateRoom(roomSetting *RoomSetting) bool {
	//TODO
	ss.Session.Info.Capacity = SessionInfo_RoomMaster
	ss.Session.AddMsgChan("RoomContent", 10)
	room := RoomManager.CreateRoom(ss.Session, roomSetting)
	ss.Session.Room = room
	ss.Session.SetState(int32(ss.StateCode) + 1)
	RoomManager.RemoveIdleUserMsgChan(ss.Session.GetMsgChan("RoomList"))
	ss.Session.CloseMsgChan("RoomList")
	ss.Session.AddMsgChan("ServerInfo", 1)
	return true
}

func (ss *UserIdleSessionState) EnterRoom(roomId int64) bool {
	room := RoomManager.Rooms[roomId]
	if room != nil {
		ss.Session.AddMsgChan("RoomContent", 2)
		if room.EnterRoom(ss.Session) {
			ss.Session.Room = room
			ss.Session.SetState(int32(ss.StateCode) + 1)
			RoomManager.RemoveIdleUserMsgChan(ss.Session.GetMsgChan("RoomList"))
			ss.Session.CloseMsgChan("RoomList")
			ss.Session.AddMsgChan("ServerInfo", 1)
			return true
		} else {
			ss.Session.CloseMsgChan("RoomContent")
		}
	}
	return false
}

type UserInRoomSessionState struct {
	SessionStateBase
}

func (ss *UserInRoomSessionState) DeleteRoom() bool {
	return false
}

func (ss *UserInRoomSessionState) ReadyRoom() bool {
	ss.Session.IsReady = true
	ss.Session.Room.CheckReady()
	return true
}

func (ss *UserInRoomSessionState) LeaveRoom() bool {
	ss.Session.IsReady = false
	ss.Session.Room = nil
	ss.Session.SetState(int32(ss.StateCode) - 1)
	return false
}

func (ss *UserInRoomSessionState) SettingCharacter(setting *CharacterSetting) bool {
	if ss.Session.User.SetCharacter(setting) {
		ss.Session.Room.UpdateRoomContent()
		return true
	}
	return false

}
func (ss *UserInRoomSessionState) CancelReady() bool {
	ss.Session.IsReady = false
	ss.Session.Room.CheckReady()
	return true
}

type ConnectingGameSessionState struct {
	SessionStateBase
}

func (ss *ConnectingGameSessionState) EndRoom() bool {
	return false
}

type sessionStateFactory struct {
}

func (sf *sessionStateFactory) makeSessionState(session *Session, state_code SessionInfo_SessionState) SessionState {
	var s SessionState
	switch state_code {
	case SessionInfo_NoSession:
		s = &NoSessionState{}
	case SessionInfo_Guest:
		s = &GuestSessionState{}
	case SessionInfo_UserIdle:
		s = &UserIdleSessionState{}
	case SessionInfo_UserInRoom:
		s = &UserInRoomSessionState{}
	case SessionInfo_ConnectingGame:
		s = &ConnectingGameSessionState{}
	default:
		s = &SessionStateBase{}
	}
	s.Lock()
	s.SetSession(session)
	s.SetStateCode(state_code)
	s.Unlock()
	return s
}

var Manager *sessionManager

var SessionStateFactory *sessionStateFactory

func init() {
	Manager = &sessionManager{
		Sessions:           make(map[int64]*Session),
		UserNameMapSession: make(map[string]int64),
	}
	SessionStateFactory = &sessionStateFactory{}
}
