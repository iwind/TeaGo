package sessions

import (
	"encoding/json"
	"fmt"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/actions"
	"github.com/iwind/TeaGo/files"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/utils/string"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileSessionManager struct {
	life       uint
	encryptKey string
	isLoading  bool
	sessionMap *sync.Map // { sid => FileSessionData, ... }
	mutex      *sync.Mutex
	dir        string // 文件存放目录
}

type FileSessionData struct {
	Sid       string            `json:"sid"`
	ExpiredAt uint              `json:"expiredAt"`
	Values    map[string]string `json:"values"`
	AccessAt  uint              `json:"accessAt"`
}

func (this *FileSessionData) fileName(dir string) string {
	return dir + Tea.DS + "session_" + this.Sid + ".data"
}

func (this *FileSessionData) isExpired() bool {
	return this.ExpiredAt < uint(time.Now().Unix())
}

func NewFileSessionManager(lifeSeconds uint, encryptKey string) *FileSessionManager {
	if lifeSeconds <= 0 {
		lifeSeconds = 86400 * 30
	}

	if len(encryptKey) > 32 {
		encryptKey = encryptKey[:32]
	} else {
		encryptKey = fmt.Sprintf("%-32s", encryptKey)
	}

	var manager = &FileSessionManager{
		life:       lifeSeconds,
		encryptKey: encryptKey,
		sessionMap: &sync.Map{},
		mutex:      &sync.Mutex{},
	}

	manager.load()

	return manager
}

func (this *FileSessionManager) Init(config *actions.SessionConfig) {
	this.life = config.Life

	// 默认为30天
	if this.life == 0 {
		this.life = 86400 * 30
	}

	if len(config.Secret) > 32 {
		this.encryptKey = config.Secret[:32]
	} else {
		this.encryptKey = fmt.Sprintf("%-32s", config.Secret)
	}

	this.sessionMap = &sync.Map{}
	this.mutex = &sync.Mutex{}

	this.load()
}

func (this *FileSessionManager) SetDir(dir string) {
	this.dir = dir
}

func (this *FileSessionManager) realDir() string {
	if len(this.dir) > 0 {
		return this.dir
	}

	return Tea.TmpDir()
}

func (this *FileSessionManager) Read(sid string) map[string]string {
	session, ok := this.sessionMap.Load(sid)
	if !ok {
		return map[string]string{}
	}
	sessionObject := session.(*FileSessionData)

	if sessionObject.isExpired() {
		return map[string]string{}
	}

	timestamp := uint(time.Now().Unix())
	sessionObject.ExpiredAt = timestamp + this.life
	sessionObject.AccessAt = timestamp

	return sessionObject.Values
}

func (this *FileSessionManager) WriteItem(sid string, key string, value string) bool {
	var values = this.Read(sid)

	this.mutex.Lock()
	values[key] = value
	this.mutex.Unlock()

	timestamp := uint(time.Now().Unix())
	session := &FileSessionData{
		Sid:       sid,
		AccessAt:  timestamp,
		ExpiredAt: timestamp + this.life,
		Values:    values,
	}

	this.sessionMap.Store(session.Sid, session)

	return this.writeSession(session)
}

func (this *FileSessionManager) writeSession(session *FileSessionData) bool {
	data, err := this.encryptData(session)
	if err != nil {
		logs.Errorf("%s", err.Error())
		return false
	}

	file := this.realDir() + Tea.DS + "session_" + session.Sid + ".data"

	this.mutex.Lock()
	defer this.mutex.Unlock()

	f := files.NewFile(file)
	err = f.Write(data)
	if err != nil {
		logs.Errorf("%s", err.Error())
		return false
	}

	return true
}

func (this *FileSessionManager) Delete(sid string) bool {
	this.sessionMap.Delete(sid)

	file := this.realDir() + Tea.DS + "session_" + sid + ".data"
	_, err := os.Stat(file)
	if err != nil {
		return true
	}

	err = os.Remove(file)
	if err != nil {
		logs.Errorf("%s", err.Error())
		return false
	}

	return true
}

func (this *FileSessionManager) encryptData(data *FileSessionData) ([]byte, error) {
	// 由于ffjson.Marshal()或json.Marshal()函数并不是并发安全的，所以需要lock
	this.mutex.Lock()
	jsonData, err := json.Marshal(data)
	this.mutex.Unlock()
	if err != nil {
		return nil, err
	}

	encryptedData, err := encrypt(jsonData, []byte(this.encryptKey))
	if err != nil {
		logs.Errorf("%s", err.Error())
		return nil, err
	}

	return encryptedData, err
}

func (this *FileSessionManager) decryptData(data []byte) (*FileSessionData, error) {
	sourceData, err := decrypt(data, []byte(this.encryptKey))
	if err != nil {
		return nil, err
	}

	var sessionData = &FileSessionData{}
	err = json.Unmarshal(sourceData, sessionData)
	if err != nil {
		return nil, err
	}

	return sessionData, nil
}

func (this *FileSessionManager) load() {
	if this.isLoading {
		return
	}

	var reg, err = stringutil.RegexpCompile("^session_(\\w+).data$")
	if err != nil {
		logs.Errorf("%s", err.Error())
		return
	}

	// 加载
	go func() {
		defer logs.Println("load sessions from files", this.realDir())

		matches, err := filepath.Glob(this.realDir() + Tea.DS + "session_*")
		if err != nil {
			logs.Errorf("%s", err.Error())
			return
		}
		for _, filename := range matches {
			regMatches := reg.FindStringSubmatch(filepath.Base(filename))
			if len(regMatches) < 1 {
				continue
			}
			sid := regMatches[1]

			data, err := os.ReadFile(filename)
			if err != nil {
				logs.Errorf("%s", err.Error())
				continue
			}

			session, err := this.decryptData(data)
			if err != nil {
				continue
			}

			// 检查sid
			if session.Sid != sid {
				this.Delete(sid)
				continue
			}

			this.sessionMap.Store(session.Sid, session)
		}
	}()

	// 更新：1分钟更新一次时间
	go func() {
		timeInterval := 60
		ticker := time.NewTicker(time.Duration(timeInterval) * time.Second)
		for {
			<-ticker.C

			this.sessionMap.Range(func(key, value interface{}) bool {
				sessionObject := value.(*FileSessionData)
				if sessionObject.isExpired() {
					this.sessionMap.Delete(key)
					os.Remove(sessionObject.fileName(this.realDir()))
				} else {
					if sessionObject.AccessAt+uint(timeInterval) > uint(time.Now().Unix()) {
						this.writeSession(sessionObject)
					}
				}

				return true
			})
		}
	}()
}
