# GKE Cleaner

An application to automate cleaning GKE clusters after a set amount of time.

There are two applications:

## GKE Cleaner

The backend that handles cleaning our GKE clusters. It works on a polling cycle
to retrieve new clusters from GKE and deleting expired clusters.

### Params

* `PORT`: The port the backend server should listen on.
* `PROJECT`: The GCP project the backend is watching.
* `GCP_SERVICE_ACCOUNT_KEY`: The GCP service account key the backend can use to
  authenticate with GCP. The key requires the GKE Cluster Admin privilege to
  both list clusters and delete clusters.
* `GCLOUD_POLL_INTERVAL`: The poll interval used to retrieve/delete clusters.
  Defaults to 10 minutes. The value must be specified in Golang's [time duration
  format](https://golang.org/pkg/time/#ParseDuration).
* `CLUSTER_LIFETIME_DURATION`: The duration of the cluster's life from the
  moment it is discovered by the backend. Defaults to 24 hours. The value must
  be specified in Golang's [time duration
  format](https://golang.org/pkg/time/#ParseDuration).
* `GCLOUD_GKE_LABEL_FILTERS`: The labels that should be used to filter the
  clusters that should be watched by the backend. If there are multiple label
  filters, they are applied independently from each other i.e it is an OR not
  AND. The value should be specified as a json array of strings. Each string
  takes the form `key=value`. For example, `["key1=value1", "key2=value2"]`.
* `BASIC_AUTH_USERNAME`: The username used for basic authentication to the REST
  API.
* `BASIC_AUTH_PASSWORD`: The password used for basic authentication to the REST
  API.
* `VCAP_SERVICES`: Used to get the credentials for the MySQL database. More info
  in [binding the DB](#binding-the-db).

### Binding the DB

The app requires a MySQL database to persist its data. To discover the database
credentials the app expects there to be a bound user provided service with a
binding name of `db`. Within the service binding, it expects to find a `uri` and
its value should be a [correctly
formatted](https://github.com/go-sql-driver/mysql#dsn-data-source-name) MySQL
connection string.


### REST API

The following endpoints exist on the backend:

* GET `/clusters`: List all known clusters.
* POST `/clusters/renew/:name`: Renews a given cluster for
  `CLUSTER_LIFETIME_DURATION`.
* POST `/clusters/ignore/:name`: Ignores a cluster i.e the cluster will NOT be
  deleted by the app.
* POST `/clusters/unignore/:name`: Unignores a previously ignored cluster i.e
  the cluster will be deleted by the app.

## UI

A very basic web ui that allows an engineer to interact with the GKE cleaner.

### Params

* `GKE_CLEANER_BACKEND_URL`: The url for the GKE cleaner backend.
* `BASIC_AUTH_USERNAME`: The username used for basic authentication to the
  webpage.
* `BASIC_AUTH_PASSWORD`: The password used for basic authentication to the
  webpage.
