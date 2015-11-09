# Intro

I like [Parse](https://parse.com) for its ease of use and I think it's really nice example about well-designed API. However it is probably best for mobile app where you don't need too much control at server side. Its query can be slow since we don't have control over index, and many services are not available like cache or websocket.

My objective for Goal is to provides ready to use implementation for CRUD, basic query, authentication and permission model, so it can be quickly setup and run. Goal relies on a few popular Go libraries like [Gorm](https://github.com/jinzhu/gorm), Gorilla [Mux](www.gorillatoolkit.org/pkg/mux) and [Sessions](www.gorillatoolkit.org/pkg/sessions).
