# GeoHash and Nearby Locations Service

This Go application provides a service to find locations within a given radius from a central point using GeoHash encoding and a trie data structure. The application includes several key functionalities, including GeoHash encoding/decoding, calculating distances using the Haversine formula, and serving a REST API to interact with location data.

## User Interface 
<img width="1511" alt="Screenshot 2024-09-30 at 6 15 21 PM" src="https://github.com/user-attachments/assets/7fbcd75c-108b-4ab8-bb77-b1a7fedf6a4e">

## Features

1. **GeoHash Encoding & Decoding**: The app encodes latitude and longitude into GeoHash strings and decodes GeoHash strings back into geographic coordinates.
2. **Trie Data Structure**: A custom trie implementation is used to store GeoHash strings and efficiently search for nearby locations.
3. **Haversine Formula**: The app uses this formula to calculate the distance between two geographical points.
4. **Find Nearby Locations**: Given a location (latitude, longitude) and a radius, the app finds locations within that radius.
5. **REST API**: A simple HTTP server provides endpoints to load location data and find nearby places.


## Key Components

### 1. Location Struct
Defines the structure for a location including the name, latitude, longitude, and distance.

```go
type Location struct {
    Name      string  `json:"name"`
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
    Distance  float64 `json:"distance"`
}
```

### 2. GeoHash Encoding and Decoding
Functions for encoding latitude and longitude into a GeoHash and decoding GeoHash strings back into coordinates.

- **EncodeGeoHash**: Converts latitude and longitude to a GeoHash string.
- **DecodeGeoHash**: Converts a GeoHash string back to latitude and longitude.


### 3. Trie Data Structure
A trie is used to store and search GeoHash strings.

- **Insert**: Inserts a GeoHash into the trie.
- **Search**: Checks if a GeoHash exists in the trie.
- **Delete**: Removes a GeoHash from the trie.
- **FindNearby**: Finds nearby GeoHashes by creating a bounding box around a location and querying for GeoHashes within the range.


### 4. Haversine Distance Calculation
The `CalculateDistance` function computes the distance between two geographic coordinates using the Haversine formula.

<img width="649" alt="Screenshot 2024-09-28 at 8 48 01 PM" src="https://github.com/user-attachments/assets/71ba6bae-a224-40be-bc9f-d2138002711e">

<img width="735" alt="Screenshot 2024-09-28 at 8 49 02 PM" src="https://github.com/user-attachments/assets/1a00fd40-0f30-440f-8d5c-6fe0869e9763">


### 5. REST API Endpoints
The application provides three main endpoints:

- **GET /load-dummy**: Loads dummy location data (based on a city code) and inserts the corresponding GeoHashes into the trie.
- **GET /find-nearby**: Finds locations within a given radius from a specified latitude and longitude.
- **GET /**: Welcome message for the GeoHash service.


### 6. Example Dummy Data (LocationData)
```go
type LocationData struct {
    Lat       float64    `json:"lat"`
    Lon       float64    `json:"lon"`
    PlaceData []Location `json:"placeData"`
}
```

## Usage

### Running the Server
1. Run the Go application to start the HTTP server:
```go
go run main.go
```

2. The server will start listening on http://localhost:8000.


### Endpoints
- **GET** `/`: Welcome message for the service.
- **GET** `/load-dummy?city=1`: Load location data for a given city (e.g., Mumbai with `city=1`).
- **GET** `/find-nearby?lat=<latitude>&lon=<longitude>&radius=<radius>`: Find places within the given radius (in kilometers) from the specified latitude and longitude.


## Example
To find nearby locations within a 10km radius of a point in Mumbai:

```shell
curl http://localhost:8000/find-nearby?lat=19.076090&lon=72.877426&radius=10
```

### Output
```json
[
    {
        "name": "Gateway of India",
        "latitude": 18.922002,
        "longitude": 72.837783,
        "distance": 6.58
    },
    {
        "name": "Marine Drive",
        "latitude": 18.920414,
        "longitude": 72.835517,
        "distance": 6.92
    }
]
```


## Conclusion
This application provides a robust and efficient way to find nearby locations using GeoHash and trie-based search. It's highly customizable and can be adapted to different cities or regions by loading different sets of location data.




