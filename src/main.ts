import { GeolocateControl, Map as MMap, Marker } from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
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
      el.innerHTML = `<svg class="marker" width="64" height="50" viewBox="0 0 278 349"><g fill="none" fill-rule="evenodd"><circle cx="139" cy="139" r="139" fill="#61AEFF"/><path fill="#FFF" fill-rule="nonzero" d="M195 227c0 6 5 11 11 11s11-5 11-11-11-22-11-22-11 16-11 22Zm-23-100h-13c-6-5-12-8-20-10v-16l-11-1-11 1v16c-8 2-15 6-20 10H56c-4 0-6 3-6 6v33c0 3 2 6 6 6h32c7 13 22 22 40 22s33-9 40-22h4c7 0 12 4 12 11 0 6 5 11 11 11h22c6 0 11-5 11-11 0-31-25-56-56-56ZM79 93l49-5 50 5c3 1 6-2 6-5V77c0-3-3-6-6-6l-39 4v-9c0-4-2-6-5-6h-11c-3 0-6 2-6 6v9l-38-4c-4 0-6 3-6 6v11c0 3 2 6 6 5Z"/><g fill="#61AEFF"><path d="M92 269c17 8 28 15 34 21 5 5 10 13 13 23v-47l-47 3ZM186 269c-17 8-28 15-34 21-5 5-10 13-13 23v-47l47 3Z"/></g><circle cx="138.5" cy="337.5" r="11.5" fill="#61AEFF"/></g></svg>`;
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
