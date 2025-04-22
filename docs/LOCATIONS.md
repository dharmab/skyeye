# Custom Locations

Server administrators can define custom locations which players can reference in ALPHA CHECK and VECTOR requests. You can define useful locations like friendly bases or fix points.

To use this feature, create a `locations.json` file. The content of the file should be a JSON array of objects. Each object should have the following properties:

- `names`: An array of strings, which are the names of the location. These names are used in the ALPHA CHECK and VECTOR requests. You can define multiple names for the same location - for example, you might have a location with both the names "Incirlik" and "Home plate". Note that the names "Bullseye" and "Tanker" are reserved and cannot be used as custom location names.
- `latitude`: A floating point number, which is the latitude of the location in decimal degrees. This should be a number between -90 and 90.
- `longitude`: A floating point number, which is the longitude of the location in decimal degrees. This should be a number between -180 and 180.

Example:

```json
[
    {
        "names": ["Incirlik", "Home plate"],
        "latitude": 37.001166662,
        "longitude": 35.422164978
    },
    {
        "names": ["Hatay", "Divert option"],
        "latitude": 36.362778,
        "longitude": 36.282222
    }
]
```

Set the path to the `locations.json` file in the `locations-file` setting in SkyEye's configuration.
