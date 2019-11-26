package pubnub

import (
	"sync"
)

//
type Listener struct {
	Status              chan *PNStatus
	Message             chan *PNMessage
	Presence            chan *PNPresence
	Signal              chan *PNMessage
	UserEvent           chan *PNUserEvent
	SpaceEvent          chan *PNSpaceEvent
	MembershipEvent     chan *PNMembershipEvent
	MessageActionsEvent chan *PNMessageActionsEvent
}

func NewListener() *Listener {
	return &Listener{
		Status:              make(chan *PNStatus),
		Message:             make(chan *PNMessage),
		Presence:            make(chan *PNPresence),
		Signal:              make(chan *PNMessage),
		UserEvent:           make(chan *PNUserEvent),
		SpaceEvent:          make(chan *PNSpaceEvent),
		MembershipEvent:     make(chan *PNMembershipEvent),
		MessageActionsEvent: make(chan *PNMessageActionsEvent),
	}
}

type ListenerManager struct {
	sync.RWMutex
	ctx                  Context
	listeners            map[*Listener]bool
	exitListener         chan bool
	exitListenerAnnounce chan bool
	pubnub               *PubNub
}

func newListenerManager(ctx Context, pn *PubNub) *ListenerManager {
	return &ListenerManager{
		listeners:            make(map[*Listener]bool, 2),
		ctx:                  ctx,
		exitListener:         make(chan bool),
		exitListenerAnnounce: make(chan bool),
		pubnub:               pn,
	}
}

func (m *ListenerManager) addListener(listener *Listener) {
	m.Lock()

	m.listeners[listener] = true
	m.Unlock()
}

func (m *ListenerManager) removeListener(listener *Listener) {
	m.pubnub.Config.Log.Println("before removeListener")
	if m.exitListener != nil {
		m.exitListener <- true
	}
	m.Lock()
	m.pubnub.Config.Log.Println("in removeListener lock")
	delete(m.listeners, listener)
	m.Unlock()
	m.pubnub.Config.Log.Println("after removeListener")
}

func (m *ListenerManager) removeAllListeners() {
	m.pubnub.Config.Log.Println("in removeAllListeners")
	m.Lock()
	lis := m.listeners
	m.Unlock()
	for l := range lis {
		delete(m.listeners, l)
	}
}

func (m *ListenerManager) copyListeners() map[*Listener]bool {
	m.RLock()
	lis := make(map[*Listener]bool)
	for k, v := range m.listeners {
		lis[k] = v
	}
	m.RUnlock()
	return lis
}

func (m *ListenerManager) announceStatus(status *PNStatus) {
	go func() {
		lis := m.copyListeners()
	AnnounceStatusLabel:
		for l := range lis {
			select {
			case <-m.exitListener:
				m.pubnub.Config.Log.Println("announceStatus exitListener")
				break AnnounceStatusLabel
			case l.Status <- status:
			}
		}
		m.pubnub.Config.Log.Println("announceStatus exit")
	}()
}

func (m *ListenerManager) announceMessage(message *PNMessage) {
	go func() {
		lis := m.copyListeners()
	AnnounceMessageLabel:
		for l := range lis {
			select {
			case <-m.exitListenerAnnounce:
				m.pubnub.Config.Log.Println("announceMessage exitListenerAnnounce")
				break AnnounceMessageLabel
			case l.Message <- message:
			}
		}

	}()
}

func (m *ListenerManager) announceSignal(message *PNMessage) {
	go func() {
		lis := m.copyListeners()

	AnnounceSignalLabel:
		for l := range lis {
			select {
			case <-m.exitListener:
				m.pubnub.Config.Log.Println("announceSignal exitListener")
				break AnnounceSignalLabel

			case l.Signal <- message:
			}
		}
	}()
}

func (m *ListenerManager) announceUserEvent(message *PNUserEvent) {
	go func() {
		lis := m.copyListeners()

	AnnounceUserEventLabel:
		for l := range lis {
			select {
			case <-m.exitListener:
				m.pubnub.Config.Log.Println("announceUserEvent exitListener")
				break AnnounceUserEventLabel

			case l.UserEvent <- message:
				m.pubnub.Config.Log.Println("l.UserEvent", message)
			}
		}
	}()
}

func (m *ListenerManager) announceSpaceEvent(message *PNSpaceEvent) {
	go func() {
		lis := m.copyListeners()

	AnnounceSpaceEventLabel:
		for l := range lis {
			m.pubnub.Config.Log.Println("l.SpaceEvent", l)
			select {
			case <-m.exitListener:
				m.pubnub.Config.Log.Println("announceSpaceEvent exitListener")
				break AnnounceSpaceEventLabel

			case l.SpaceEvent <- message:
				m.pubnub.Config.Log.Println("l.SpaceEvent", message)
			}
		}
	}()
}

func (m *ListenerManager) announceMembershipEvent(message *PNMembershipEvent) {
	go func() {
		lis := m.copyListeners()

	AnnounceMembershipEvent:
		for l := range lis {
			select {
			case <-m.exitListener:
				m.pubnub.Config.Log.Println("announceMembershipEvent exitListener")
				break AnnounceMembershipEvent

			case l.MembershipEvent <- message:
				m.pubnub.Config.Log.Println("l.MembershipEvent", message)
			}
		}
	}()
}

func (m *ListenerManager) announceMessageActionsEvent(message *PNMessageActionsEvent) {
	go func() {
		lis := m.copyListeners()

	AnnounceMessageActionsEvent:
		for l := range lis {
			select {
			case <-m.exitListener:
				m.pubnub.Config.Log.Println("announceMessageActionsEvent exitListener")
				break AnnounceMessageActionsEvent

			case l.MessageActionsEvent <- message:
				m.pubnub.Config.Log.Println("l.MessageActionsEvent", message)
			}
		}
	}()
}

func (m *ListenerManager) announcePresence(presence *PNPresence) {
	go func() {
		lis := m.copyListeners()

	AnnouncePresenceLabel:
		for l := range lis {
			select {
			case <-m.exitListener:
				m.pubnub.Config.Log.Println("announcePresence exitListener")
				break AnnouncePresenceLabel

			case l.Presence <- presence:
			}
		}
	}()
}

// PNStatus is the status struct
type PNStatus struct {
	Category              StatusCategory
	Operation             OperationType
	ErrorData             error
	Error                 bool
	TLSEnabled            bool
	StatusCode            int
	UUID                  string
	AuthKey               string
	Origin                string
	ClientRequest         interface{} // Should be same for non-google environment
	AffectedChannels      []string
	AffectedChannelGroups []string
}

// PNMessage is the Message Response for Subscribe
type PNMessage struct {
	Message           interface{}
	UserMetadata      interface{}
	SubscribedChannel string
	ActualChannel     string
	Channel           string
	Subscription      string
	Publisher         string
	Timetoken         int64
}

// PNPresence is the Message Response for Presence
type PNPresence struct {
	Event             string
	UUID              string
	SubscribedChannel string
	ActualChannel     string
	Channel           string
	Subscription      string
	Occupancy         int
	Timetoken         int64
	Timestamp         int64
	UserMetadata      map[string]interface{}
	State             interface{}
	Join              []string
	Leave             []string
	Timeout           []string
	HereNowRefresh    bool
}

// PNUserEvent is the Response for an User Event
type PNUserEvent struct {
	Event             PNObjectsEvent
	UserID            string
	Description       string
	Timestamp         string
	Name              string
	ExternalID        string
	ProfileURL        string
	Email             string
	Created           string
	Updated           string
	ETag              string
	Custom            map[string]interface{}
	SubscribedChannel string
	ActualChannel     string
	Channel           string
	Subscription      string
}

// PNSpaceEvent is the Response for a Space Event
type PNSpaceEvent struct {
	Event             PNObjectsEvent
	SpaceID           string
	Description       string
	Timestamp         string
	Name              string
	Created           string
	Updated           string
	ETag              string
	Custom            map[string]interface{}
	SubscribedChannel string
	ActualChannel     string
	Channel           string
	Subscription      string
}

// PNMembershipEvent is the Response for a Membership Event
type PNMembershipEvent struct {
	Event             PNObjectsEvent
	UserID            string
	SpaceID           string
	Description       string
	Timestamp         string
	Custom            map[string]interface{}
	SubscribedChannel string
	ActualChannel     string
	Channel           string
	Subscription      string
}

// PNMessageActionsEvent is the Response for a Message Actions Event
type PNMessageActionsEvent struct {
	Event             PNMessageActionsEventType
	Data              PNMessageActionsResponse
	SubscribedChannel string
	ActualChannel     string
	Channel           string
	Subscription      string
}
