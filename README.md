# Greenlight

This application is a JSON api that is can be used for retrieving and managing information about movies  

It has public and private endpoints that serves the necessary task to accompolish the tasks of this application

In sumary it contains the following *endpoints* and their corrospoinding *actions*.

##### Greenlight endpoints and their actions

No |Method        | URL Pattern     |Action                
-- |------------  | ----------------|----------------------
  1 |GET           | /v1/healthcheck | Displays system info and status of the applicatoin
  2 |GET           | /v1/movies      | Shows the details of all movies  
  3 |POST          | /v1/movies      | Creates a new movies
  4 |GET           | /v1/movies/:id  | Shows the details of a movie specified by "id"
  5 |PATCH         | /v1/movies/:id  | Updates the details of a movie specied by "id"
  6 |DELETE        | /v1/movies/:id  | Deletes the movies with id= "id"
  7 |POST          | /v1/users       | Registers a new user
  8 |PUT           | /v1/users/activated| Activate a specific user
  9 |POST          | /v1/tokens/authentication| generates new authenticatin token
  10|/v1/metrics   | /v1/metrics | Displays the application runtime metrics

  ##### Step by step demonstration and usage of each end point is shown  bellow


 1. ##### Display system ifo and application status

 To get the system info of the application status, just make a GET request to the 
 ``` console
 /v1/healthcheck
 ```
end point.
Since our applicaiton is hosted at the address **https://167.71.254.102/**, our request should be as follows.
###### Request 
``` console
$ curl -k https://167.71.254.102/v1/healthcheck
```
###### Response
``` json
{
	"status": "available",
	"system_info": {
		"environment": "production",
		"version": "v1.0.1-0-ge8569f9"
	}
}
```
The -k parameter here is used to allow insecure server connections when using SSL.

2. ##### Get the details of all movies

###### Request

``` console 
$ curl -k https://167.71.254.102/v1/movies
```
###### Response
``` json
{
	"error": "you must be authenticated to access this resource"
}
```
This is becuase only an authenticated can access the details of the movies.

To access this endpoint let's first register a test.

3. ##### Register a user

###### Request
``` console 
$ BODY='{"name": "user1", "email":"user1@example.com", "password": "pa55wordtest"}'
```
``` console 
$ curl -k -i -d "$BODY" -X POST https://167.71.254.102/v1/v1/users
```

###### Response

```
HTTP/2 202 </br>
content-type: application/json</br>
date: Wed, 27 Oct 2021 08:30:54 GMT </br>
server: Caddy</br>
vary: Origin</br>
vary: Access-Control-Request-Method</br>
vary: Authorization</br>
content-length: 146
```

``` json
{
	"user": {
		"id": 5,
		"created_at": "2021-10-27T04:30:55-04:00",
		"name": "user1",
		"email": "user1@example.com",
		"activated": false
	}
}
```

3. ##### Activating the registered user
Since the application is setup with SMTP credentials for my Mailtrap inbox, an activation email will be sent to the registered users email as follows.

------------------------------

<p>Hi user1,</p>
<p>Thanks for signing up for a Greenlight account. We're excited to have you on board!</p>
<p>For future reference, your user ID number is 5.</p>
<p>Please send a request to the <code>PUT /v1/users/activated</code> endpoint with the
following JSON body to activate your account:</p>
<pre><code>
{"token": "FE6QMEYH7LOYJT55MOCSKFTOP4"}
</code></pre>
<p>Please note that this is a one-time use token and it will expire in 1 day.</p>
<p>Thanks,</p>
<p>The Greenlight Team</p>

--------------------------------

Now let's activate the user using the token sent in the email.

###### Request

``` console 
$ curl -k -i -d '{"token": "FE6QMEYH7LOYJT55MOCSKFTOP4"}' -X PUT https://167.71.254.102/v1/users/activated
```

###### Response
```
HTTP/2 200
content-type: application/json
date: Wed, 27 Oct 2021 08:53:09 GMT
server: Caddy
vary: Origin
vary: Access-Control-Request-Method
vary: Authorization
content-length: 145
```

```json
{
	"user": {
		"id": 5,
		"created_at": "2021-10-27T04:30:55-04:00",
		"name": "user1",
		"email": "user1@example.com",
		"activated": true
	}
}
```
As we can see it in the response json the user has been activated succesfully. Now let's get authenticated as this user and then we can access the details of all the movies or a specific a movie.

4. ##### Authenticating the user

###### Request

``` console
$ curl -k -d '{"email":"user1@example.com", "password": "pa55wordtest"}' -X POST 'https://167.71.254.102/v1/tokens/authentication'

```

###### Response

``` json
{
	"authentication_token": {
		"token": "ZNKKSMRBXEZGL3GZ3EKUOXG3JI",
		"expiry": "2021-10-28T05:21:23.081283942-04:00"
	}
}

```
##### Get access of the movies the database
* ###### Get details of all movies
   *Request*


``` console
$ curl -k -H "Authorization: Bearer ZNKKSMRBXEZGL3GZ3EKUOXG3JI" https://167.71.254.102/v1/movies

```
  *Response*

``` json
{
	"metadata": {
		"current_page": 1,
		"page_size": 20,
		"first_page": 1,
		"last_page": 2,
		"total_records": 24
	},
	"movies": [
		{
			"id": 1,
			"title": "Moana",
			"year": 2016,
			"runtime": "107 mins",
			"genres": [
				"animation",
				"adventure"
			],
			"version": 1
		},
		{
			"id": 2,
			"title": "Black Panther",
			"year": 2018,
			"runtime": "134 mins",
			"genres": [
				"sci-fi",
				"action",
				"adventure"
			],
			"version": 1
		},
		{
			"id": 3,
			"title": "Deadpool",
			"year": 2016,
			"runtime": "108 mins",
			"genres": [
				"action",
				"comedy"
			],
			"version": 1
		},
		{
			"id": 4,
			"title": "The Breakfast Club",
			"year": 1986,
			"runtime": "96 mins",
			"genres": [
				"drama"
			],
			"version": 1
		},
		{
			"id": 5,
			"title": "The Graduate",
			"year": 1967,
			"runtime": "106 mins",
			"genres": [
				"Comedy",
				"Drama",
				"Romance"
			],
			"version": 1
		},
		{
			"id": 6,
			"title": "The Shawshank Redemption",
			"year": 1994,
			"runtime": "142 mins",
			"genres": [
				"Crime",
				"Drama"
			],
			"version": 1
		},
		{
			"id": 7,
			"title": "Crocodile Dundee",
			"year": 1986,
			"runtime": "97 mins",
			"genres": [
				"Adventure",
				"Comedy"
			],
			"version": 1
		},
		{
			"id": 8,
			"title": "Valkyrie",
			"year": 2008,
			"runtime": "121 mins",
			"genres": [
				"Drama",
				"History",
				"Thriller"
			],
			"version": 1
		},
		{
			"id": 9,
			"title": "City of God",
			"year": 2002,
			"runtime": "130 mins",
			"genres": [
				"Crime",
				"Drama"
			],
			"version": 1
		},
		{
			"id": 10,
			"title": "Memento",
			"year": 2000,
			"runtime": "113 mins",
			"genres": [
				"Mystery",
				"Thriller"
			],
			"version": 1
		},
		{
			"id": 11,
			"title": "Stardust",
			"year": 2007,
			"runtime": "127 mins",
			"genres": [
				"Adventure",
				"Family",
				"Fantasy"
			],
			"version": 1
		},
		{
			"id": 12,
			"title": "Apocalypto",
			"year": 2006,
			"runtime": "139 mins",
			"genres": [
				"Action",
				"Adventure",
				"Drama"
			],
			"version": 1
		},
		{
			"id": 13,
			"title": "Taxi Driver",
			"year": 1976,
			"runtime": "113 mins",
			"genres": [
				"Crime",
				"Drama"
			],
			"version": 1
		},
		{
			"id": 14,
			"title": "No Country for Old Men",
			"year": 2007,
			"runtime": "122 mins",
			"genres": [
				"Crime",
				"Drama",
				"Thriller"
			],
			"version": 1
		},
		{
			"id": 15,
			"title": "Planet 51",
			"year": 2009,
			"runtime": "91 mins",
			"genres": [
				"Animation",
				"Adventure",
				"Comedy"
			],
			"version": 1
		},
		{
			"id": 16,
			"title": "Looper",
			"year": 2012,
			"runtime": "119 mins",
			"genres": [
				"Action",
				"Crime",
				"Drama"
			],
			"version": 1
		},
		{
			"id": 17,
			"title": "Corpse Bride",
			"year": 2005,
			"runtime": "77 mins",
			"genres": [
				"Animation",
				"Drama",
				"Family"
			],
			"version": 1
		},
		{
			"id": 18,
			"title": "The Third Man",
			"year": 1949,
			"runtime": "93 mins",
			"genres": [
				"Film-Noir",
				"Mystery",
				"Thriller"
			],
			"version": 1
		},
		{
			"id": 19,
			"title": "The Beach",
			"year": 2000,
			"runtime": "119 mins",
			"genres": [
				"Adventure",
				"Drama",
				"Romance"
			],
			"version": 1
		},
		{
			"id": 20,
			"title": "Scarface",
			"year": 1983,
			"runtime": "170 mins",
			"genres": [
				"Crime",
				"Drama"
			],
			"version": 1
		}
	]
}
```


* ###### Do a search by movie title
    *Request*
``` console
$ curl -k -H "Authorization: Bearer ZNKKSMRBXEZGL3GZ3EKUOXG3JI" 'https://167.71.254.102/v1/movies?title=the+club'

```
  *Response*

``` json
{
	"metadata": {
		"current_page": 1,
		"page_size": 20,
		"first_page": 1,
		"last_page": 1,
		"total_records": 1
	},
	"movies": [
		{
			"id": 4,
			"title": "The Breakfast Club",
			"year": 1986,
			"runtime": "96 mins",
			"genres": [
				"drama"
			],
			"version": 1
		}
	]
}

```

* ###### Do a search by movie genre
    *Request*
``` console
$ curl -k -H "Authorization: Bearer ZNKKSMRBXEZGL3GZ3EKUOXG3JI" 'https://167.71.254.102/v1/movies?genres=Crime,Drama'
```

   *Response*
``` json
{
	"metadata": {
		"current_page": 1,
		"page_size": 20,
		"first_page": 1,
		"last_page": 1,
		"total_records": 6
	},
	"movies": [
		{
			"id": 6,
			"title": "The Shawshank Redemption",
			"year": 1994,
			"runtime": "142 mins",
			"genres": [
				"Crime",
				"Drama"
			],
			"version": 1
		},
		{
			"id": 9,
			"title": "City of God",
			"year": 2002,
			"runtime": "130 mins",
			"genres": [
				"Crime",
				"Drama"
			],
			"version": 1
		},
		{
			"id": 13,
			"title": "Taxi Driver",
			"year": 1976,
			"runtime": "113 mins",
			"genres": [
				"Crime",
				"Drama"
			],
			"version": 1
		},
		{
			"id": 14,
			"title": "No Country for Old Men",
			"year": 2007,
			"runtime": "122 mins",
			"genres": [
				"Crime",
				"Drama",
				"Thriller"
			],
			"version": 1
		},
		{
			"id": 16,
			"title": "Looper",
			"year": 2012,
			"runtime": "119 mins",
			"genres": [
				"Action",
				"Crime",
				"Drama"
			],
			"version": 1
		},
		{
			"id": 20,
			"title": "Scarface",
			"year": 1983,
			"runtime": "170 mins",
			"genres": [
				"Crime",
				"Drama"
			],
			"version": 1
		}
	]
}

```

 ###### Get a paginated lists

   *Request*

``` console
$ curl -k -H "Authorization: Bearer ZNKKSMRBXEZGL3GZ3EKUOXG3JI" 'https://167.71.254.102/v1/movies?page_size=5&page=3' 
```
   *Response*
``` json

{
	"metadata": {
		"current_page": 1,
		"page_size": 5,
		"first_page": 1,
		"last_page": 5,
		"total_records": 24
	},
	"movies": [
		{
			"id": 1,
			"title": "Moana",
			"year": 2016,
			"runtime": "107 mins",
			"genres": [
				"animation",
				"adventure"
			],
			"version": 1
		},
		{
			"id": 2,
			"title": "Black Panther",
			"year": 2018,
			"runtime": "134 mins",
			"genres": [
				"sci-fi",
				"action",
				"adventure"
			],
			"version": 1
		},
		{
			"id": 3,
			"title": "Deadpool",
			"year": 2016,
			"runtime": "108 mins",
			"genres": [
				"action",
				"comedy"
			],
			"version": 1
		},
		{
			"id": 4,
			"title": "The Breakfast Club",
			"year": 1986,
			"runtime": "96 mins",
			"genres": [
				"drama"
			],
			"version": 1
		},
		{
			"id": 5,
			"title": "The Graduate",
			"year": 1967,
			"runtime": "106 mins",
			"genres": [
				"Comedy",
				"Drama",
				"Romance"
			],
			"version": 1
		}
	]
}

```

###### Get a sorted lists

Lets get some movies sorted by title in descending order

   *Request*
``` console

curl -k -H "Authorization: Bearer ZNKKSMRBXEZGL3GZ3EKUOXG3JI" 'https://167.71.254.102/v1/movies?page_size=3&page=5&sort=-title'

```

   *Response*
``` json
{
	"metadata": {
		"current_page": 5,
		"page_size": 3,
		"first_page": 1,
		"last_page": 8,
		"total_records": 24
	},
	"movies": [
		{
			"id": 14,
			"title": "No Country for Old Men",
			"year": 2007,
			"runtime": "122 mins",
			"genres": [
				"Crime",
				"Drama",
				"Thriller"
			],
			"version": 1
		},
		{
			"id": 1,
			"title": "Moana",
			"year": 2016,
			"runtime": "107 mins",
			"genres": [
				"animation",
				"adventure"
			],
			"version": 1
		},
		{
			"id": 10,
			"title": "Memento",
			"year": 2000,
			"runtime": "113 mins",
			"genres": [
				"Mystery",
				"Thriller"
			],
			"version": 1
		}
	]
}

```

Adding and updating a movie item needs a user with a  write and read permisions. By default any registered and activated user has a read permission. However for a user to have a write permission it must be granted by the database admin.




For example the user identified by "user1@example.com" has only "read" permission as seen bellow.

![db info](screenshot1.png)

To do write operations on the movies table, one should be signed in as admin@greenlight.com

Let's just do that right now.

###### Get authenticated as admin@greenlight

   *Request*
```console 
$ curl -k -d '{"email": "admin@greenlight.com", "password": "fakepasswordhere"}' -X POST 'https://167.71.254.102/v1/tokens/authentication'
```
   *Response*

   ``` json
   {
	"authentication_token": {
		"token": "H4NURVCDHKKEOHIRJA3TEIAV34",
		"expiry": "2021-10-28T08:41:28.068293831-04:00"
	}
}
   ```
Now let's Add, Update and Delete a movie using the authentication token of the admin user.

4. ##### Add a new movie 

  *Request*

``` console
$ BODY='{"title": "The Deer Hunter","year": 1978,"runtime": "183 mins","genres": ["War"]}'
```


``` console
$ curl -k -H "Authorization: Bearer H4NURVCDHKKEOHIRJA3TEIAV34" -d "$BODY" -X POST 'https://167.71.254.102/v1/movies'
```

 *Response*

 ``` json
 {
	"movie": {
		"id": 25,
		"title": "The Deer Hunter",
		"year": 1978,
		"runtime": "183 mins",
		"genres": [
			"War"
		],
		"version": 1
	}
}

 ```

 The new movie has been succesfully added.

 Now let's update the genres field 

 5. ##### Updating a movie

 

add "Darama" to the genres array and update the movie 

 *Request*

 ``` console
$ curl -k -H "Authorization: Bearer H4NURVCDHKKEOHIRJA3TEIAV34" -d '{"genres": ["War","Drama"]}' -X PATCH 'https://167.71.254.102/v1/movies/25'
 ```
*Response*

``` json
{
	"movie": {
		"id": 25,
		"title": "The Deer Hunter",
		"year": 1978,
		"runtime": "183 mins",
		"genres": [
			"War",
			"Drama"
		],
		"version": 2
	}
}
```
Notice the "version" field of the movie object. It shows the number of edits or updates made.

6. ##### Delete a Movie by id

Let's delete the last added movie 

 *Request*

 ``` console
 $ curl -k -H "Authorization: Bearer H4NURVCDHKKEOHIRJA3TEIAV34" -X DELETE 'https://167.71.254.102/v1/movies/25'
 ```

 *Response*

 ``` json

 {
	"message": "movie successfully deleted"
}
```
Let's try accessing the movie

 *Request*

 ``` console
 curl -k -H "Authorization: Bearer H4NURVCDHKKEOHIRJA3TEIAV34" -X GET 'https://167.71.254.102/v1/movies/25'
 ```

  *Response*

 ``` json
 
 {
	"error": "the requested resource could not be found"
}

 ```

 This tells us the movie has been deleted.




