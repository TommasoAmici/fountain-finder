import { GeolocateControl, LngLat, LngLatLike, Map as MMap, Marker } from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import markerSVG from "./marker.svg?raw";
import "./style.css";

navigator.geolocation.getCurrentPosition(async position => {
  const map = new MMap({
    container: "map",
    style: `https://api.maptiler.com/maps/streets/style.json?key=${
      import.meta.env.VITE_MAP_STYLE_KEY
    }`,
    hash: true,
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

  const markers = new Map<number, Marker>();
  let prevPosition: LngLat | null = null;

  // Find fountains and add them to the map
  const addMarkers = async () => {
    const bounds = map.getBounds();

    // if the current position is not too far from the previous
    // position we can return early
    if (prevPosition !== null && bounds.contains(prevPosition)) {
      return;
    }

    // extend bounds to have some buffer zone all around
    const startLat = Math.min(bounds._ne.lat, bounds._sw.lat);
    const endLat = Math.max(bounds._ne.lat, bounds._sw.lat);
    const startLng = Math.min(bounds._ne.lng, bounds._sw.lng);
    const endLng = Math.max(bounds._ne.lng, bounds._sw.lng);
    const extendBy = 0.1;
    bounds.extend([startLng - extendBy, startLat - extendBy]);
    bounds.extend([endLng + extendBy, endLat + extendBy]);

    // remove markers off screen
    markers.forEach((m, k) => {
      if (!bounds.contains(m.getLngLat())) {
        m.remove();
        markers.delete(k);
      }
    });

    const precision = 4;
    const res = await fetch(
      `/api/fountains/${startLng.toFixed(precision)}/${startLat.toFixed(
        precision
      )}/${endLng.toFixed(precision)}/${endLat.toFixed(precision)}`
    );
    const fountains: OverpassResponse = await res.json();

    fountains.elements.forEach(fountain => {
      const lngLat: LngLatLike = [fountain.lon, fountain.lat];
      if (markers.get(fountain.id) === undefined && bounds.contains(lngLat)) {
        const el = document.createElement("div");
        el.innerHTML = markerSVG;
        const marker = new Marker({ element: el, anchor: "bottom" })
          .setLngLat(lngLat)
          .addTo(map);
        markers.set(fountain.id, marker);
      }
    });
    prevPosition = map.getCenter();
  };

  map.on("load", () => {
    geolocate.trigger();
    addMarkers();
  });
  map.on("moveend", () => {
    addMarkers();
  });
});
