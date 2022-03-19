import { GeolocateControl, Map as MMap, Marker } from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import markerSVG from "./marker.svg?raw";
import { findFountains } from "./osm";
import "./style.css";

const options = {
  container: "map",
  style: `https://api.maptiler.com/maps/streets/style.json?key=${
    import.meta.env.VITE_MAP_STYLE_KEY
  }`,
};

const uniqueInArray = (arr: any[]) => [
  ...new Map(arr.map(item => [item["id"], item])).values(),
];

navigator.geolocation.getCurrentPosition(async position => {
  const map = new MMap({
    ...options,
    center: [position.coords.longitude, position.coords.latitude],
    zoom: 16,
  });
  // Initialize the geolocate control.
  const geolocate = new GeolocateControl({
    fitBoundsOptions: {
      maxZoom: 16,
    },
    positionOptions: {
      enableHighAccuracy: true,
    },
    trackUserLocation: true,
  });
  map.addControl(geolocate);

  let fountains: Fountain[] = [];
  // Find fountains and add them to the map
  const addMarkers = async () => {
    const bounds = map.getBounds();
    const newFountains = await findFountains(bounds);
    fountains = uniqueInArray([...fountains, ...newFountains]);
    fountains.forEach(fountain => {
      const el = document.createElement("div");
      el.innerHTML = markerSVG;
      new Marker(el).setLngLat([fountain.lon, fountain.lat]).addTo(map);
    });
  };

  map.on("load", function () {
    geolocate.trigger();
    addMarkers();
  });
  map.on("moveend", function () {
    addMarkers();
  });
});
