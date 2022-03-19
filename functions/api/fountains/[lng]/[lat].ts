interface Fountain {
  type: "node";
  id: number;
  lat: number;
  lon: number;
  tags: {
    amenity: string;
    operator: string;
    "operator:wikidata"?: string;
    "operator:wikipedia"?: string;
    bottle?: "yes" | "no";
  };
}
interface OverpassResponse {
  version: number;
  generator: string;
  osm3s: {
    timestamp_osm_base: string;
    copyright: string;
  };
  elements: Fountain[];
}

export const onRequestGet: PagesFunction = async ({ params }) => {
  console.log(params);
  let lat = 0;
  let lng = 0;
  if (typeof params.lng === "string") {
    lng = parseInt(params.lng);
  } else {
    lat = parseInt(params.lng[0]);
  }
  if (typeof params.lat === "string") {
    lat = parseInt(params.lat);
  } else {
    lat = parseInt(params.lat[0]);
  }

  const bounds = `(${lat},${lng},${lat + 1},${lng + 1})`;
  const query = encodeURIComponent(
    `[out:json][timeout:25];(node["amenity"="drinking_water"]${bounds};);out body;>;out skel qt;`
  );
  const url = `https://overpass.kumi.systems/api/interpreter/?data=${query}`;
  let response = await fetch(url, {
    cf: {
      cacheTtl: 3600,
      cacheEverything: true,
    },
  });
  const data: OverpassResponse = await response.json();

  const coordinates = data.elements.map(e => [e.lat, e.lon]);

  return new Response(JSON.stringify(coordinates), {
    headers: {
      "content-type": "application/json;charset=UTF-8",
      "Cache-Control": "max-age=3600",
    },
  });
};
