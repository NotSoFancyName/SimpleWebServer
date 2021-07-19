# SimpleWebServer

To start the server run:

```
$ docker-compose up 
```

It will start server at the ```localhost:8081```. 
The server has only one REST API endpoint ```GET /api/block/<block_number>/total```, where ***block_number*** is an unsigned integer.

You could test it by using ```curl```, for example:

```
$ curl http://localhost:8081/api/block/1155089/total
```
