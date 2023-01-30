## ORM 框架

### 一.  什么是 orm 框架？

**对象关系映射（Object Relational Mapping，简称ORM）模式是一种为了解决面向对象与关系数据库存在的互不匹配的现象的技术**。ORM框架是连接数据库的桥梁，只要提供了持久化类与表的映射关系，ORM框架在运行时就能参照映射文件的信息，把对象持久化到数据库中。

## 二. 为什么使用ORM?

在最开始没有 orm 的时候，开发者们在开发过程种与数据库的交互常常避免不了以下步骤：

- **手写 SQL**：
- **问题**： 
  - 容易出错
  - 难以重构
- **手动处理结果集** 
  - **问题**： 
    - 过多的样板代码
    - 开发者的将过多的精力投入到跟业务没关系的地方

**如单纯的使用 database/sql 操作数据库如下**：

```go
type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

func Open(driver string, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	db, err := Open("mysql", "root:123456@tcp(localhost:3306)/integration_test")
	if err != nil {
		panic(err)
	}
	rows, err := db.QueryContext(context.Background(), "SELECT `id`,`first_name` FROM `test_model` LIMIT 1;")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		tm := &TestModel{}
		err = rows.Scan(&tm.Id, &tm.FirstName)
		if err != nil {
			panic(err)
		}
		fmt.Println(tm)
	}
}
```

所以我们就希望有这么一个东西， 帮助我们完成这两个步骤，这也就 是 ORM 诞生的初衷。

### 三. ORM 框架主要职责

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675045880475-5c4507b4-55cb-49e8-a84d-a69a65fc4212.png#averageHue=%23fafafa&clientId=uefbb0f2a-aa8d-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=646&id=u1fba56c5&margin=%5Bobject%20Object%5D&name=image.png&originHeight=807&originWidth=1389&originalType=binary&ratio=1&rotation=0&showTitle=false&size=106508&status=done&style=none&taskId=u643d2e37-2e41-4fbd-be66-c0e7659c36b&title=&width=1111.2)


-  **对象 -> SQL**
   当用户输入一个对象的时候，能够产生对应的 SQL。比如在插入场景下，一个 User 对象应该生成一条 INSERT INTO 语句 
-  **结果集  -> 对象**
   当收到数据库返回的行时，能 够将行组装成对象，并返回给用户。例如一条 SELECT 语句查出来一条数据，并将该数据组装 成一个 User 对象返回给用户 

### 四. GO 主流的 ORM 框架

#### 4.1 Beego orm

-  **入门例子** 

```go
package beego
import (
	"github.com/beego/beego/v2/client/orm"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

// User -
type User struct {
	ID   int    `orm:"column(id)"`
	Name string `orm:"column(name)"`
}

// 注册模型、驱动以及 DB
func init() {
	// need to register models in init
	orm.RegisterModel(new(User))

	// need to register db driver
	orm.RegisterDriver("sqlite3", orm.DRSqlite)

	// need to register default database
	orm.RegisterDataBase("default",
		"sqlite3", "beego.db")
}

// 创建 ORM 实例
func TestCRUD(t *testing.T) {
	// automatically build table
	orm.RunSyncdb("default", false, true)

	// create orm object
	o := orm.NewOrm()

	// data
	user := new(User)
	user.Name = "mike"

	// insert data
	o.Insert(user)
}
```


-  **源码解析** 
   -  元数据
      元数据是对模型的描述， 在 Beego 里分成了  modelInfo -> fields -> fieldInfo 三个层级

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675046064200-16880a35-e687-4b75-b3d2-94b44dc9f992.png#averageHue=%232c2c2b&clientId=uefbb0f2a-aa8d-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=402&id=uf223149e&margin=%5Bobject%20Object%5D&name=image.png&originHeight=503&originWidth=1262&originalType=binary&ratio=1&rotation=0&showTitle=false&size=74546&status=done&style=none&taskId=uf509a37e-860c-429a-9b57-e4767b257f7&title=&width=1009.6)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675046080273-0e31ee62-c04a-4ee1-b3f8-7d01231fa471.png#averageHue=%232d2c2b&clientId=uefbb0f2a-aa8d-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=507&id=u4e59eebb&margin=%5Bobject%20Object%5D&name=image.png&originHeight=634&originWidth=1253&originalType=binary&ratio=1&rotation=0&showTitle=false&size=138923&status=done&style=none&taskId=u77112e76-b207-4bb7-9679-0ac7660e36a&title=&width=1002.4)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675046092650-59f29264-ffb6-4b24-a6e3-aa199083b7e0.png#averageHue=%232c2c2b&clientId=uefbb0f2a-aa8d-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=694&id=u3c1168cf&margin=%5Bobject%20Object%5D&name=image.png&originHeight=867&originWidth=1290&originalType=binary&ratio=1&rotation=0&showTitle=false&size=148065&status=done&style=none&taskId=u0deb025f-398e-4fb0-a980-1e0de7664b7&title=&width=1032)

   -  查询接口

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675046120629-36dd3879-1f62-44dd-bf70-54b5770a5bd2.png#averageHue=%232e2d2c&clientId=uefbb0f2a-aa8d-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=667&id=u2aac7431&margin=%5Bobject%20Object%5D&name=image.png&originHeight=834&originWidth=1270&originalType=binary&ratio=1&rotation=0&showTitle=false&size=165403&status=done&style=none&taskId=ue91a7e7a-48eb-4a9d-8e6e-1866267847f&title=&width=1016)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675046139049-f1cf7a4c-e382-41d8-ab59-e229f89b0ade.png#averageHue=%232d2c2b&clientId=uefbb0f2a-aa8d-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=738&id=u14d6d117&margin=%5Bobject%20Object%5D&name=image.png&originHeight=922&originWidth=1242&originalType=binary&ratio=1&rotation=0&showTitle=false&size=268373&status=done&style=none&taskId=ub0e1baf7-8f17-4052-af7a-4036567b639&title=&width=993.6)

   -  事务接口 
      -  事务接口分成两类： 
         -  普通的 Begin、Commit 和 Rollback 
         -  闭包形式的 DoXX

![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675046161071-49b8048c-c261-44b0-aeeb-fced50d44934.png#averageHue=%23fbfbfa&clientId=uefbb0f2a-aa8d-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=459&id=u7909b056&margin=%5Bobject%20Object%5D&name=image.png&originHeight=574&originWidth=1280&originalType=binary&ratio=1&rotation=0&showTitle=false&size=47994&status=done&style=none&taskId=uffb88854-8fa1-4301-969d-1b8514c7528&title=&width=1024)
![image.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675046179454-2397a4e4-de51-48c8-8a75-bc3115cb153b.png#averageHue=%232d2c2b&clientId=uefbb0f2a-aa8d-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=570&id=u51779602&margin=%5Bobject%20Object%5D&name=image.png&originHeight=712&originWidth=1291&originalType=binary&ratio=1&rotation=0&showTitle=false&size=190373&status=done&style=none&taskId=ucf84d777-d893-45ee-8016-088b8dfd7a8&title=&width=1032.8)

#### 4.2 Gorm

-  **入门例子** 

```go
package gorm

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

type Product struct {
	gorm.Model
	Code  string `gorm:"column(code)"`
	Price uint
}

func  (p Product) TableName() string {
	return "product_t"
}

func (p *Product) BeforeSave(tx *gorm.DB) (err error) {
	println("before save")
	return
}

func (p *Product) AfterSave(tx *gorm.DB) (err error) {
	println("after save")
	return
}

func (p *Product) BeforeCreate(tx *gorm.DB) (err error) {
	println("before create")
	return
}

func (p *Product) AfterCreate(tx *gorm.DB) (err error) {
	println("after create")
	// 刷新缓存
	return
}

func (p *Product) BeforeUpdate(tx *gorm.DB) (err error) {
	println("before update")
	return
}

func (p *Product) AfterUpdate(tx *gorm.DB) (err error) {
	println("after update")
	// 刷新缓存
	return
}

func (p *Product) BeforeDelete(tx *gorm.DB) (err error) {
	println("before update")
	return
}

func (p *Product) AfterDelete(tx *gorm.DB) (err error) {
	println("after update")
	return
}

func  (p *Product) AfterFind(tx *gorm.DB) (err error) {
	println("after find")
	return
}

func TestCRUD(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// 打印 SQL，但不执行
	db.DryRun = true

	// Migrate the schema
	db.AutoMigrate(&Product{})

	// Create
	db.Create(&Product{Code: "D42", Price: 100})


	// Read
	var product Product
	db.First(&product, 1) // find product with integer primary key
	db.First(&product, "code = ?", "D42") // find product with code D42

	// Update - update product's price to 200
	db.Model(&product).Update("Price", 200)
	// Update - update multiple fields
	db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
	db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete - delete product
	db.Delete(&product, 1)
}
```


-  **源码解析** 
   -  元数据 

![image-20230103155858220.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675046277860-0a490a74-4e65-489b-86dd-74bff6ec9848.png#averageHue=%232c2c2b&clientId=uefbb0f2a-aa8d-4&crop=0&crop=0&crop=1&crop=1&from=drop&id=u752cfbd3&margin=%5Bobject%20Object%5D&name=image-20230103155858220.png&originHeight=863&originWidth=1041&originalType=binary&ratio=1&rotation=0&showTitle=false&size=111857&status=done&style=none&taskId=uf05e884d-c1db-4a2c-83c4-a63b860fc62&title=)
![image-20230103155919092.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675046299599-07a30d1a-1508-4340-b07d-426557888288.png#averageHue=%232c2c2b&clientId=uefbb0f2a-aa8d-4&crop=0&crop=0&crop=1&crop=1&from=drop&id=u579f4a12&margin=%5Bobject%20Object%5D&name=image-20230103155919092.png&originHeight=873&originWidth=903&originalType=binary&ratio=1&rotation=0&showTitle=false&size=90587&status=done&style=none&taskId=u9f70d927-67a0-43a3-af88-590cefb3ab8&title=)

      -  GORM 的元数据看起来和 Beego ORM元数据差不多，很难理解结构体里每一个字段的作用。 一个比较大的不同是 GORM 只有 Schame-> Field 两级。

   -  查询接口

![image-20230103160206374.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675046360832-2f06eb3c-e9f1-492d-b352-52c7c89c61ae.png#averageHue=%232c2c2b&clientId=uefbb0f2a-aa8d-4&crop=0&crop=0&crop=1&crop=1&from=drop&id=uf54ac9a4&margin=%5Bobject%20Object%5D&name=image-20230103160206374.png&originHeight=276&originWidth=1035&originalType=binary&ratio=1&rotation=0&showTitle=false&size=20677&status=done&style=none&taskId=u180e2ae6-b359-4a9e-85ea-978cc33f860&title=)
![image-20230103160242253.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675046374673-223d375f-c088-4e02-a2f2-9d8fd0fc2868.png#averageHue=%232d2d2c&clientId=uefbb0f2a-aa8d-4&crop=0&crop=0&crop=1&crop=1&from=drop&id=u1faff699&margin=%5Bobject%20Object%5D&name=image-20230103160242253.png&originHeight=300&originWidth=903&originalType=binary&ratio=1&rotation=0&showTitle=false&size=36881&status=done&style=none&taskId=ua5ad5ccc-81c9-4eba-adcb-3241481cece&title=)
![image-20230103160313641.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675046454736-0b943915-9409-4403-a9c0-e6a5d5a464d5.png#averageHue=%232d2c2b&clientId=uefbb0f2a-aa8d-4&crop=0&crop=0&crop=1&crop=1&from=drop&id=udb05a206&margin=%5Bobject%20Object%5D&name=image-20230103160313641.png&originHeight=681&originWidth=856&originalType=binary&ratio=1&rotation=0&showTitle=false&size=71055&status=done&style=none&taskId=u0203c1df-011b-494d-9527-aa3dbf04fe3&title=)

   -  事务接口 
      - Begin、Commit、Rollback
      - 闭包接口
      - save![image-20230103160753900.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675046471762-f593cbb9-3214-4fcd-a1a4-64c693fda48b.png#averageHue=%232c2c2b&clientId=uefbb0f2a-aa8d-4&crop=0&crop=0&crop=1&crop=1&from=drop&id=u6b9c0908&margin=%5Bobject%20Object%5D&name=image-20230103160753900.png&originHeight=581&originWidth=1170&originalType=binary&ratio=1&rotation=0&showTitle=false&size=71697&status=done&style=none&taskId=uca63a2b1-31b3-45a3-a836-9b1912d3d3b&title=) point 支持


#### 4.3 Ent

Ent 和 Beego ORM、GORM 有一个设计理念上 的本质区别，就是 Ent 采用的是代码生成技术
![image-20230103163431209.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675046502133-5cf991fd-b4f9-44e5-b600-36a933fc6108.png#averageHue=%23dcc58b&clientId=uefbb0f2a-aa8d-4&crop=0&crop=0&crop=1&crop=1&from=drop&id=ube105be8&margin=%5Bobject%20Object%5D&name=image-20230103163431209.png&originHeight=785&originWidth=1414&originalType=binary&ratio=1&rotation=0&showTitle=false&size=78411&status=done&style=none&taskId=u741a38c3-ea72-4de9-9c19-6236a1df355&title=)

-  **入门例子** 

```go
package ent

import (
	"context"
	"entgo.io/ent/dialect"
	"gitee.com/geektime-geekbang/geektime-go/orm/ent/ent"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"testing"
)

func TestEntCURD(t *testing.T) {
	// Create an ent.Client with in-memory SQLite database.
	client, err := ent.Open(dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()
	ctx := context.Background()
	// Run the automatic migration tool to create all schema resources.
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	err = client.User.Create().Exec(context.Background())
}
```


### 五. ORM 核心功能模块

![image-20230103144621802.png](https://cdn.nlark.com/yuque/0/2023/png/27674489/1675046517885-51517f73-0e7b-4c12-bab4-8b169b859326.png#averageHue=%23fafafa&clientId=uefbb0f2a-aa8d-4&crop=0&crop=0&crop=1&crop=1&from=drop&id=ue9094c4c&margin=%5Bobject%20Object%5D&name=image-20230103144621802.png&originHeight=891&originWidth=1212&originalType=binary&ratio=1&rotation=0&showTitle=false&size=95672&status=done&style=none&taskId=ub65423ff-4f91-4d38-80e0-0bfa8396435&title=)

- **SQL**：必须要支持的就是增删改查，DDL 一般是作为 一个扩展功能，或者作为一个工具来提供。
- **映射**：将结果集封装成对象，性能瓶颈。
- **事务**：主要在于维护好事务状态。
- **元数据**：SQL 和映射两个部分的基石。
- **AOP**：处理横向关注点。
- **关联关系**：部分 ORM 框架会提供，性价比低。
- **方言**：兼容不同的数据库，至少要兼容 MySQL、 SQLite、 PostgreSQL。

### 六. 总结

- **ORM 框架的核心是什么**？**SQL 构造**和**处理结果集**。在别的语言里面，因为底层库可能不够强大， ORM 框架还要解决连接和会话管理的问题，Go 是不需要的。
- **ORM 是什么**？ORM 是指**对象关系映射**，一般是**指用于语言对象和关系型数据库行相互转化的工具**。
- **为什么要使用 ORM 框架**？本质上来说，不使用 ORM 框架也是可以的，但我们就要花费很多时间在处理拼接 SQL、处理返回的行上。这些本身是不难但很琐碎的事情，所以我们需要一个 ORM 框架来帮我们解决这两个问题。
- **ORM 的优点**？ORM 的优点就是 API 对编程更加友好，开发效率更高。
- **使用 ORM 性能会更好吗**？显然不能，直接写 SQL 的性能最好。
- **ORM 怎么使用缓存**？利用 AOP 来拦截到查询，而后嵌入缓存。