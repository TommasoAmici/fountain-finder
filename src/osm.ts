import { LngLatBounds } from "maplibre-gl";

const BASE_URL = import.meta.env.VITE_OSM_BASE_URL;

export const findFountains = async (bounds: LngLatBounds) => {
  const query = `[out:json][timeout:25];(node["amenity"="drinking_water"](${bounds._sw.lat},${bounds._sw.lng},${bounds._ne.lat},${bounds._ne.lng}););out body;>;out skel qt;`;
  const url = `${BASE_URL}/interpreter/?data=${encodeURIComponent(query)}`;
  const res = await fetch(url, { headers: { "content-type": "application/json" } });
  const data: OverpassResponse = await res.json();
  return data.elements;
};
