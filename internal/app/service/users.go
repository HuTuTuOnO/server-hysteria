package service

import (
	"fmt"
	"github.com/xflash-panda/server-hysteria/internal/pkg/api"
)
import "github.com/xflash-panda/server-hysteria/internal/pkg/counter"
import "github.com/xtls/xray-core/common/task"
import log "github.com/sirupsen/logrus"
import "sync"
import "time"

type Config struct {
	FetchUserInterval     time.Duration
	ReportTrafficInterval time.Duration
}

type UsersService struct {
	client         *api.Client
	config         *Config
	userManager    *UserManager
	trafficManager *TrafficManager
	userList       *[]api.UserInfo
	fuPeriodicTask *task.Periodic
	rtPeriodicTask *task.Periodic
}

func NewUsersService(config *Config, client *api.Client) *UsersService {
	return &UsersService{client: client, config: config, userManager: newUserManager(), trafficManager: newTrafficManager()}
}

func (s *UsersService) Init() error {
	userList, err := s.client.GetUserList()
	if err != nil {
		return err
	}
	s.userList = userList
	s.userManager.addUsers(*userList)
	log.Infof("Added %d new users", len(*userList))
	return nil
}

func (s *UsersService) Start() error {
	s.fuPeriodicTask = &task.Periodic{
		Interval: s.config.FetchUserInterval,
		Execute:  s.FetchUsersTask,
	}

	s.rtPeriodicTask = &task.Periodic{
		Interval: s.config.ReportTrafficInterval,
		Execute:  s.ReportTrafficsTask,
	}

	log.Infoln("Start fetch users task")
	err := s.fuPeriodicTask.Start()
	if err != nil {
		return fmt.Errorf("start fetch users erorr:%s", err)
	}
	log.Infoln("Start report traffic task")
	err = s.rtPeriodicTask.Start()
	if err != nil {
		return fmt.Errorf("start report traffic erorr:%s", err)
	}
	return nil
}

func (s *UsersService) Close() error {
	if err := s.fuPeriodicTask.Close(); err != nil {
		log.Warn("fetch task close error: ", err)
	}
	if err := s.rtPeriodicTask.Close(); err != nil {
		log.Warn("report task close error: ", err)
	}
	return nil
}

func (s *UsersService) FetchUsersTask() error {
	// Update User
	newUserList, err := s.client.GetUserList()
	if err != nil {
		log.Errorln(err)
		return nil
	}

	deleted, added := s.compareUserList(newUserList)
	if len(added) > 0 {
		s.userManager.addUsers(added)
	}

	if len(deleted) > 0 {
		s.userManager.deleteUsers(deleted)
	}
	log.Infof("%d user deleted, %d user added", len(deleted), len(added))
	log.Infof("current users: %d", s.userManager.countUsers())
	s.userList = newUserList
	return nil
}

func (s *UsersService) toUserTraffics() []*api.UserTraffic {
	return s.trafficManager.toUserTraffics()
}

func (s *UsersService) ReportTrafficsTask() error {
	userTraffics := s.toUserTraffics()
	log.Infof("%d user traffic needs to be reported", len(userTraffics))
	if len(userTraffics) > 0 {
		err := s.client.ReportUserTraffic(userTraffics)
		if err != nil {
			log.Errorln(err)
			return nil
		}
		s.trafficManager.clear()
	}
	return nil
}

func (s *UsersService) Auth(uuid string) (int, bool) {
	return s.userManager.auth(uuid)
}

func (s *UsersService) compareUserList(newUsers *[]api.UserInfo) (deleted, added []api.UserInfo) {
	msrc := make(map[api.UserInfo]byte) //按源数组建索引
	mall := make(map[api.UserInfo]byte) //源+目所有元素建索引

	var set []api.UserInfo //交集

	//1.源数组建立map
	for _, v := range *s.userList {
		msrc[v] = 0
		mall[v] = 0
	}
	//2.目数组中，存不进去，即重复元素，所有存不进去的集合就是并集
	for _, v := range *newUsers {
		l := len(mall)
		mall[v] = 1
		if l != len(mall) { //长度变化，即可以存
			l = len(mall)
		} else { //存不了，进并集
			set = append(set, v)
		}
	}
	//3.遍历交集，在并集中找，找到就从并集中删，删完后就是补集（即并-交=所有变化的元素）
	for _, v := range set {
		delete(mall, v)
	}
	//4.此时，mall是补集，所有元素去源中找，找到就是删除的，找不到的必定能在目数组中找到，即新加的
	for v := range mall {
		_, exist := msrc[v]
		if exist {
			deleted = append(deleted, v)
		} else {
			added = append(added, v)
		}
	}

	return deleted, added
}

func (s *UsersService) GetTrafficItem(userId int) *TrafficItem {
	item := s.trafficManager.load(userId)
	if item == nil {
		newItem := newTrafficItem()
		s.trafficManager.set(userId, newItem)
		return newItem
	}
	return item
}

type UserManager struct {
	store sync.Map
}

func newUserManager() *UserManager {
	return &UserManager{store: sync.Map{}}
}

func (um *UserManager) addUsers(users []api.UserInfo) {
	for _, user := range users {
		um.store.Store(user.UUID, user.ID)
	}
}

func (um *UserManager) deleteUsers(users []api.UserInfo) {
	for _, user := range users {
		log.Infoln("--DELETE", user.UUID)
		um.store.Delete(user.UUID)
	}
}

func (um *UserManager) countUsers() int {
	length := 0
	um.store.Range(func(_, _ interface{}) bool {
		length++
		return true
	})
	return length
}

func (um *UserManager) auth(uuid string) (int, bool) {
	userId, ok := um.store.Load(uuid)
	if !ok {
		return -1, false
	}
	return userId.(int), ok
}

type TrafficManager struct {
	store sync.Map
}

func (tm *TrafficManager) toUserTraffics() []*api.UserTraffic {
	userTraffics := make([]*api.UserTraffic, 0)
	tm.store.Range(func(key, value any) bool {
		userId := key.(int)
		trafficItem := value.(*TrafficItem)
		if trafficItem.Up.Value() > 0 || trafficItem.Down.Value() > 0 || trafficItem.Count.Value() > 0 {
			userTraffics = append(userTraffics, &api.UserTraffic{
				UID:      userId,
				Upload:   trafficItem.Up.Value(),
				Download: trafficItem.Down.Value(),
				Count:    trafficItem.Count.Value(),
			})
		}
		return true
	})
	return userTraffics
}

func (tm *TrafficManager) load(userId int) *TrafficItem {
	if item, ok := tm.store.Load(userId); !ok {
		return nil
	} else {
		return item.(*TrafficItem)
	}
}

func (tm *TrafficManager) set(userId int, item *TrafficItem) {
	tm.store.Store(userId, item)
}

func (tm *TrafficManager) forRange(f func(key, value any) bool) {
	tm.store.Range(f)
}

func (tm *TrafficManager) delete(userId int) {
	tm.store.Delete(userId)
}

func (tm *TrafficManager) clear() {
	tm.store.Range(func(key interface{}, value interface{}) bool {
		tm.store.Delete(key)
		return true
	})
}

func newTrafficManager() *TrafficManager {
	return &TrafficManager{store: sync.Map{}}
}

type TrafficItem struct {
	Up    *counter.Counter
	Down  *counter.Counter
	Count *counter.Counter
}

func (t *TrafficItem) delete() {
	t.Count.Reset()
	t.Down.Reset()
	t.Count.Reset()
}

func newTrafficItem() *TrafficItem {
	return &TrafficItem{counter.NewCounter(0), counter.NewCounter(0), counter.NewCounter(0)}
}
