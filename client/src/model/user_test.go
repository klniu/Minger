package model

import (
	"golang.org/x/crypto/bcrypt"
	. "gopkg.in/check.v1"
)

type UserSuite struct{}

var _ = Suite(&UserSuite{})

func (s *UserSuite) TestCryptPass(c *C) {
	user := User{UserName: "admin", Password: "f865b53623b121fd34ee5426c792e5c33af8c227",
		Email: "example@example.org", Role: Admin}
	c.Assert(user.cryptPass(), IsNil)
	c.Assert(bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("f865b53623b121fd34ee5426c792e5c33af8c227")), IsNil)
}
func (s *UserSuite) TestCreateAdminUser(c *C) {
	// CreateAdminUser is included in emptyDB
	d := emptyDB()
	var user User
	c.Assert(d.First(&user).Error, IsNil)
	c.Assert(user.UserName, Equals, "admin")
	c.Assert(bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("f865b53623b121fd34ee5426c792e5c33af8c227")), IsNil)
}

func (s *UserSuite) TestCreateUser(c *C) {
	d := emptyDB()
	user := User{UserName: "admin", Password: "f865b53623b121fd34ee5426c792e5c33af8c227", Email: "example@example.org", Role: Admin}
	c.Assert(d.CreateUser(user), ErrorMatches, "创建用户失败")
}

func (s *UserSuite) TestLogin(c *C) {
	d := emptyDB()
	user, err := d.Login("admin", "f865b53623b121fd34ee5426c792e5c33af8c227")
	c.Assert(err, IsNil)
	c.Assert(user.ID, Equals, 1)
	c.Assert(user.UserName, Equals, "admin")

	_, err = d.Login("admin", "admin1234")
	c.Assert(err, ErrorMatches, "密码错误")

	_, err = d.Login("admin123", "admin1234")
	c.Assert(err, ErrorMatches, "用户不存在")

	_, err = d.Login(" ", "admin1234")
	c.Assert(err, ErrorMatches, "用户名不能为空")
}

func (s *UserSuite) TestChangePassword(c *C) {
	d := emptyDB()
	err := d.ChangePassword(1, "f865b53623b121fd34ee5426c792e5c33af8c227", "7b902e6ff1db9f560443f2048974fd7d386975b0")
	c.Assert(err, IsNil)
	var user User
	c.Assert(d.First(&user).Error, IsNil)
	c.Assert(bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("7b902e6ff1db9f560443f2048974fd7d386975b0")), IsNil)

	err = d.ChangePassword(1, "wrongoriginpass", "7b902e6ff1db9f560443f2048974fd7d386975b0")
	c.Assert(err, ErrorMatches, "原始密码错误")

	err = d.ChangePassword(2, "wrongoriginpass", "7b902e6ff1db9f560443f2048974fd7d386975b0")
	c.Assert(err, ErrorMatches, "用户不存在")
}
