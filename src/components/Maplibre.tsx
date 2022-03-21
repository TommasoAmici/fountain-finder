import {
  GeolocateControl,
  LngLat,
  LngLatBounds,
  LngLatLike,
  Map as MMap,
  Marker,
} from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import { Component, createSignal, onCleanup, onMount } from "solid-js";
import markerSVG from "./marker.svg?raw";

const MAX_ZOOM = 16;
const MAX_MARKERS = 200;

/**
 * Extends bounds to have a buffer zone all around
 * @param bounds
 */
const addBufferZone = (bounds: LngLatBounds) => {
  const extendBy = 0.001;
  bounds.extend([bounds.getWest() - extendBy, bounds.getSouth() - extendBy]);
  bounds.extend([bounds.getEast() + extendBy, bounds.getNorth() + extendBy]);
};

const cacheKey = (bounds: LngLatBounds) => {
  const startLat = Math.trunc(bounds.getSouth());
  const startLng = Math.trunc(bounds.getWest());
  return `${startLng}-${startLat}`;
};

/**
 * Fetch all water fountains in `bounds` from Overpass API
 * @param bounds
 * @returns a response from Overpass API
 */
const getFountainsInBounds = async (
  cachedCoords: Map<string, boolean>,
  bounds: LngLatBounds
): Promise<OverpassResponse> => {
  let n, e, s, w;
  if (cachedCoords.get(cacheKey(bounds))) {
    s = Math.trunc(bounds.getSouth());
    n = s + 1;
    w = Math.trunc(bounds.getWest());
    e = w + 1;
  } else {
    w = bounds.getWest().toFixed(2);
    s = bounds.getSouth().toFixed(2);
    e = bounds.getEast().toFixed(2);
    n = bounds.getNorth().toFixed(2);
  }
  const res = await fetch(`/api/fountains/${w}/${s}/${e}/${n}`);
  return res.json();
};

const Maplibre: Component = () => {
  let mapContainer: HTMLDivElement | undefined = undefined;

  const [getMap, setMap] = createSignal<MMap>();
  const [prevCenter, setPrevCenter] = createSignal<LngLat | null>(null);

  const markers = new Map<number, Marker>();
  const cacheHits = new Map<string, boolean>();

  onMount(() => {
    if (mapContainer !== undefined) {
      const geolocate = new GeolocateControl({
        fitBoundsOptions: { maxZoom: MAX_ZOOM },
        positionOptions: { enableHighAccuracy: true },
        trackUserLocation: true,
      });

      const newMap = new MMap({
        container: mapContainer,
        // style: "https://demotiles.maplibre.org/style.json",
        style: `https://api.maptiler.com/maps/streets/style.json?key=${
          import.meta.env.VITE_MAP_STYLE_KEY
        }`,
        hash: true,
        zoom: MAX_ZOOM,
        pixelRatio: Math.min(window.devicePixelRatio, 2),
      })
        .addControl(geolocate)
        .on("load", () => {
          geolocate.trigger();
          addMarkers();
        })
        .on("moveend", addMarkers);
      setMap(newMap);
    }
  });

  onCleanup(() => {
    getMap()?.remove();
  });

  const recordCacheHit = (bounds: LngLatBounds) => {
    cacheHits.set(cacheKey(bounds), true);
  };

  const addMarkers = async () => {
    const map = getMap();
    if (map === undefined) return;

    const bounds = map.getBounds();
    addBufferZone(bounds);

    // if the current position is not too far from the previous
    // position we can return early
    const prev = prevCenter();
    if (prev !== null && bounds.contains(prev)) {
      return;
    }

    const fountains = await getFountainsInBounds(cacheHits, bounds);
    if (fountains.cache === true) {
      recordCacheHit(bounds);
    }

    fountains.elements.forEach(fountain => {
      const lngLat: LngLatLike = [fountain.lon, fountain.lat];
      if (markers.get(fountain.id) === undefined) {
        if (bounds.contains(lngLat) || markers.size < MAX_MARKERS) {
          const el = document.createElement("div");
          el.innerHTML = markerSVG;
          const marker = new Marker({ element: el, anchor: "bottom" })
            .setLngLat(lngLat)
            .addTo(map);
          markers.set(fountain.id, marker);
        }
      }
    });
    setPrevCenter(map.getCenter());
  };

  return <div class="h-full w-full" ref={mapContainer} />;
};

export default Maplibre;
