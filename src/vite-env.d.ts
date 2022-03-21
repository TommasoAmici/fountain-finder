/// <reference types="vite/client" />
interface ImportMetaEnv {
  readonly VITE_MAP_STYLE_KEY: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

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
  cache?: boolean;
}
