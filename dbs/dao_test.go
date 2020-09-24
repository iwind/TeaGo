package dbs

import (
	"encoding/json"
	"fmt"
	"github.com/iwind/TeaGo/maps"
	"log"
	"testing"
	"time"
)

type UserDAO DAO

type User struct {
	Id         int    `field:"id"`
	Gender     int    `field:"gender"`
	Birthday   string `field:"birthday"`
	Name       string `field:"name"`
	CreatedAt  int    `field:"created_at"`
	UpdatedAt  int    `field:"updated_at"`
	State      int    `field:"state"`
	IsShop     bool   `field:"is_shop"`
	CountViews int    `field:"count_views"`
}

type UserOperator struct {
	Id         interface{}
	Gender     interface{}
	Birthday   interface{}
	Name       interface{}
	CreatedAt  interface{}
	UpdatedAt  interface{}
	State      interface{}
	IsShop     interface{}
	CountViews interface{}
}

func NewUserDAO() *UserDAO {
	return NewDAO(&UserDAO{
		DAOObject{
			DB:     "db2",
			Table:  "pp_users",
			Model:  new(User),
			PkName: "id",
		},
	}).(*UserDAO)
}

func (dao *UserDAO) FindUsers() ([]maps.Map, []string, error) {
	return dao.Query().
		Where("id>=:minId").
		Param("minId", 1).
		Limit(10).
		Desc("id").
		//Attr("id", 1).
		//Attr("is_shop", false).
		//Attr("state", 1).
		//Debug(true).
		FindOnes()
}

func (dao *UserDAO) FindUser(userId int) (*User, error) {
	var value, err = dao.Find(userId)
	if err != nil {
		return nil, err
	}
	return value.(*User), nil
}

func (dao *UserDAO) CreateUser(name string, gender int) (int64, error) {
	var user = new(User)
	user.Name = name
	user.Gender = gender
	user.State = 1

	dao.Save(user)

	return int64(user.Id), nil
}

func TestDAOQuery(t *testing.T) {
	var dao = NewUserDAO()

	t.Logf("%p, %p, %p, %p", NewUserDAO(), NewUserDAO(), dao.Query(), dao.Query())

	var ones, _, err = dao.FindUsers()
	if err != nil {
		t.Fatal(err)
	}
	for _, one := range ones {
		oneBytes, err := json.Marshal(one)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(fmt.Sprintf("%d: %s", one["id"], string(oneBytes)))
	}

}

func TestDAOFind(t *testing.T) {
	var dao = NewUserDAO()
	t.Log(dao.FindUser(1024))
	t.Log(dao.Find(1025))
	t.Log(dao.Exist(1026))
	t.Log(dao.Exist(10086))
}

func TestDaoDelete(t *testing.T) {
	var dao = NewUserDAO()
	t.Log(dao.Delete(8273))
}

func TestDaoSave(t *testing.T) {
	var dao = NewUserDAO()

	now := time.Now()

	for i := 0; i < 10; i++ {
		var user = new(UserOperator)
		//user.Id = 19
		user.Name = "李白6"
		user.Gender = 2
		user.Birthday = "1996-10-20"
		//user.State = 0
		user.IsShop = false
		//user.CountViews = SQL("count_views+1")

		_, err := dao.Save(user)
		if err != nil {
			log.Fatal(err)
		}
	}

	//t.Log("Id:", user.Id, "Old:", user, "New:", newUser)
	t.Log(float64(time.Since(now).Nanoseconds()) / 1000000)
}

func TestDaoSaveEmpty(t *testing.T) {
	type UserOperator struct {
		Id   interface{}
		Name interface{}
		Age  interface{}
	}
	var user = new(UserOperator)
	t.Log(user)
	user.Id = 1
	user.Name = FuncRand()
	t.Log(user)
	t.Log("user.Name == nil", user.Name == nil)
	t.Log("user.Age == nil", user.Age == nil)
}

type User2 struct {
	Gender DAOInt `field:"gender"`
	IsShop *bool  `field:"is_shop"`
}

type DAOInt int

func TestDaoSave2(t *testing.T) {
	var user2 = new(User2)
	user2.Gender = 1
}

func TestDAOObject_NotifyInsert(t *testing.T) {
	dao := &DAOObject{}
	dao.OnInsert(func() error {
		t.Log("func1")
		return nil
	})
	dao.OnInsert(func() error {
		t.Log("func2")
		return ErrNotFound
	})
	dao.OnInsert(func() error {
		t.Log("func3")
		return nil
	})
	err := dao.NotifyInsert()
	t.Log("expected:", err)
}
