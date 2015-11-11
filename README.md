# Intro

I like [Parse](https://parse.com) for its ease of use and I think it's really nice example about well-designed API. However it is probably best for mobile backend where you don't need too much control at server side. Its query can be slow since we don't have control over index, and many services are not available like cache or websocket.

My objective for Goal is to provide ready to use implementation for CRUD, basic query, authentication and permission model, so it can be quickly setup and run. Goal relies on a few popular Go libraries like [Gorm](https://github.com/jinzhu/gorm), Gorilla [Mux](www.gorillatoolkit.org/pkg/mux) and [Sessions](www.gorillatoolkit.org/pkg/sessions).

# Setup Model and database connection

Since Goal relies on Gorm, you can just define your models as described in `Gorm` documentation. For example:

```go
type testuser struct {
	ID       uint `gorm:"primary_key"`
	Username string
	Password string
	Name     string
	Age      int
}

type article struct {
	ID     uint `gorm:"primary_key"`
	Author *testuser
	Title  string
	Read   string
	Write  string
}

func setupDB()  {
  var err error
  db, err = gorm.Open("sqlite3", ":memory:")
  if err != nil {
    panic(err)
  }

  db.SingularTable(true)

  // Setup database
  goal.InitGormDb(&db)
}
```


# Setup basic CRUD and Query

Goal simplifies many ideas from [Sleepy Framework](https://github.com/dougblack/sleepy), where each model defines methods to handle basic CRUD and Query for that model. Goal provides ready-to-use implementation for each of these method. If you don't want to support any method in your API, just do not implement it, Goal will return 405 HTTP error code to the client.

```go
// Define HTTP methods to support
func (user *testuser) Get(w http.ResponseWriter, request *http.Request) (int, interface{}, error) {
	return goal.Read(user, request)
}

func (user *testuser) Post(w http.ResponseWriter, request *http.Request) (int, interface{}, error) {
	return goal.Create(user, request)
}

func (user *testuser) Put(w http.ResponseWriter, request *http.Request) (int, interface{}, error) {
	return goal.Update(user, request)
}

func (user *testuser) Delete(w http.ResponseWriter, request *http.Request) (int, interface{}, error) {
	return goal.Delete(user, request)
}
```

Goal uses Gorilla Mux to route the request correctly to the handler. You can use all features of Gorilla Mux with `api.Mux()`

```go
func main() {
  // Initialize API
  api := goal.NewAPI()

  // Add paths to correct model
  user := &testuser{}
  db.AutoMigrate(user)
	api.AddCrudResource(user, "/testuser", "/testuser/{id:[a-zA-Z0-9]+}")
	api.AddQueryPath(user, "/query/testuser/{query}")

  http.Handle("/", r)
  http.ListenAndServe(":8080", api.Mux())
}
```

Goal predefines a set of default paths based on the model's table name (as defined by Gorm):

```go
// Extract name of resource type
name := TableName(resource)

// Default path to interact with resource
createPath := fmt.Sprintf("/%s", name)
detailPath := fmt.Sprintf("/%s/{id:[a-zA-Z0-9]+}", name)

// Query path
queryPath := fmt.Sprintf("/query/%s/{query}", TableName(resource))
```

So if you want to quickly setup your API with default paths, use below methods:

```go
models := []interface{}{&testuser{}, &article{}}

// Add default path
for _, model := range models {
  goal.RegisterModel(model)
}
```

# Caching

Goal supports caching to quickly retrieve data, and also includes basic implementation for Redis. If you have setup Redis in your server, use it like below:

```go
var (
	redisAddress   = flag.String("redis-address", ":6379", "Address to the Redis server")
	maxConnections = flag.Int("max-connections", 10, "Max connections to Redis")
)

func SetupRedis() {
	pool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", *redisAddress)

		if err != nil {
			return nil, err
		}

		return c, err
	}, *maxConnections)

	redisCache := &goal.RedisCache{}
	err = redisCache.InitRedisPool(pool)
	if err == nil {
		goal.RegisterCacher(redisCache)
	}
}
```

`redisCache` is just instance of `goal.Cacher` interface. By calling `goal.RegisterCacher`, goal can use the cacher to quickly get and set your data into cache. If you use Memcached or other type of cache, just implement Cacher interface for your respective cache and register it with Goal.

# Authentication

Goal uses Gorilla Session to support user authentication. First you need to let Goal know which model represents your user:

```go
goal.SetUserModel(&testuser{})
```

You can also uses Goal default paths for routing authentication requests, or change it if you like.

```go
// AddDefaultAuthPaths route request to the model which implement
// authentications
func (api *API) AddDefaultAuthPaths(resource interface{}) {
	api.Mux().Handle("/auth/register", api.registerHandler(resource))
	api.Mux().Handle("/auth/login", api.loginHandler(resource))
	api.Mux().Handle("/auth/logout", api.logoutHandler(resource))
}
```

Now in the model that you represent your user, implement `goal.Registerer`, `goal.Loginer` and `goal.Logouter` interface. As usual, Goal provides basic implementations for these use cases.

```go
// Setup methods to conform to auth interfaces
func (user *testuser) Register(w http.ResponseWriter, req *http.Request) (int, interface{}, error) {
	currentUser, err := goal.RegisterWithPassword(w, req, "username", "password")

	if err != nil {
		return 500, nil, err
	}

	return 200, currentUser, nil
}

func (user *testuser) Login(w http.ResponseWriter, req *http.Request) (int, interface{}, error) {
	currentUser, err := goal.LoginWithPassword(w, req, "username", "password")

	if err != nil {
		return 500, nil, err
	}

	return 200, currentUser, nil
}

func (user *testuser) Logout(w http.ResponseWriter, req *http.Request) (int, interface{}, error) {
	goal.HandleLogout(w, req)
	return 200, nil, nil
}
```

You can utilize above implementations or roll out your own authentication mechanism, for example login with Facebook/Google etc. To properly set request/response session, use `SetUserSession(w, request, user)`. After user authenticated successfully, you can retrieve current user by `GetCurrentUser(request)`

# Access Controls

Goal defines simple system based on roles to guard your record. First your user model needs to implement `goal.Roler` interface, so Goal knows which role current request has:

```go
// Satisfy Roler interface
func (user *testuser) Role() []string {
	ownRole := fmt.Sprintf("testuser:%v", user.ID)
	roles := []string{ownRole}

	return roles
}
```

For your record you want to protect, implements `goal.PermitWriter` and `goal.PermitReader` interface:

```go
func (art *article) PermitRead() []string {
	return strings.Split(art.Read, ",")
}

func (art *article) PermitWrite() []string {
	return strings.Split(art.Write, ",")
}
```

If a record doesn't implement any Permit* interfaces above, Goal assumes it can be accessed by public

# License

MIT License
