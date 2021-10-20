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

  Step by step demonstration and usage of each end point is shown  as bellow
