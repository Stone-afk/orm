# orm

#### 介绍
简单的自建 go orm 框架, 该框架主要目的是学习并深入了解 orm 框架的设计, 参考 beego orm, gorm以及eorm

#### 软件架构
![image-20230103133624088](/docs/images/image-20230103133624088.png)

#### 核心功能
- **SQL**：必须要支持的就是增删改查。
- **映射**：将结果集封装成对象，性能瓶颈。
- **事务**：主要在于维护好事务状态。
- **元数据**：SQL 和映射两个部分的基石。
- **AOP**：处理横向关注点。
- **方言**：兼容不同的数据库，至少要兼容 MySQL、 SQLite。

