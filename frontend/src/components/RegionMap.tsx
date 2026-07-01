"use client";

import { useEffect, useRef } from "react";
import {
  Map as MapLibreMap,
  NavigationControl,
  Popup,
  type GeoJSONSource,
  type MapLayerMouseEvent,
  type StyleSpecification,
} from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import { AnimatePresence, motion } from "framer-motion";
import type { Feature, FeatureCollection, Geometry, Position } from "geojson";
import countriesData from "@/data/countries.json";
import { useSearch } from "@/store/useSearch";
import { useHover } from "@/store/useHover";
import { CONTACT_CHIPS, contactHref } from "@/lib/contacts";
import type { Company, Region } from "@/lib/types";

/** Convert companies to a GeoJSON point collection for the markers layer.
 *  `key` doubles as the feature id (via promoteId) so the results list can
 *  highlight a marker on hover. */
function toPoints(companies: Company[]): FeatureCollection {
  return {
    type: "FeatureCollection",
    features: companies
      .filter((c) => c.lat !== 0 || c.lon !== 0)
      .map((c) => ({
        type: "Feature",
        geometry: { type: "Point", coordinates: [c.lon, c.lat] },
        properties: {
          key: `${c.osmType}/${c.osmId}`,
          name: c.name || "—",
          category: c.subcategory || c.category || "",
          website: c.website || "",
          phone: c.phone || "",
          email: c.email || "",
          telegram: c.telegram || "",
          whatsapp: c.whatsapp || "",
          instagram: c.instagram || "",
          facebook: c.facebook || "",
          vk: c.vk || "",
        },
      })),
  };
}

const countries = countriesData as unknown as FeatureCollection;

/** A free OpenStreetMap raster basemap — no API key, ODbL attribution. */
const baseStyle: StyleSpecification = {
  version: 8,
  sources: {
    osm: {
      type: "raster",
      tiles: ["https://tile.openstreetmap.org/{z}/{x}/{y}.png"],
      tileSize: 256,
      maxzoom: 19,
      attribution: "© OpenStreetMap contributors",
    },
  },
  layers: [{ id: "osm", type: "raster", source: "osm" }],
};

/** Walk every coordinate pair in a geometry, expanding a running bbox. */
function expandBBox(geometry: Geometry, box: [number, number, number, number]): void {
  const visit = (coords: Position | Position[] | Position[][] | Position[][][]): void => {
    if (typeof coords[0] === "number") {
      const [lon, lat] = coords as Position;
      if (lon < box[0]) box[0] = lon;
      if (lat < box[1]) box[1] = lat;
      if (lon > box[2]) box[2] = lon;
      if (lat > box[3]) box[3] = lat;
      return;
    }
    for (const c of coords as Position[]) {
      visit(c as unknown as Position);
    }
  };

  if (geometry.type === "GeometryCollection") {
    for (const g of geometry.geometries) expandBBox(g, box);
    return;
  }
  visit(geometry.coordinates as Position[]);
}

/** Compute [minLon, minLat, maxLon, maxLat] for a feature's geometry. */
function bboxOf(feature: Feature): [number, number, number, number] {
  const box: [number, number, number, number] = [
    Infinity,
    Infinity,
    -Infinity,
    -Infinity,
  ];
  if (feature.geometry) expandBBox(feature.geometry, box);
  return box;
}

export function RegionMap({
  fullScreen = false,
  companies = [],
}: {
  fullScreen?: boolean;
  companies?: Company[];
}) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<MapLibreMap | null>(null);
  const hoveredRef = useRef<string | number | null>(null);
  const selectedRef = useRef<string | number | null>(null);
  const readyRef = useRef(false);

  const setRegion = useSearch((s) => s.setRegion);
  const region = useSearch((s) => s.region);

  // Keep the latest setter without re-initialising the map on each render.
  const setRegionRef = useRef(setRegion);
  useEffect(() => {
    setRegionRef.current = setRegion;
  }, [setRegion]);

  // Hold the latest companies so the map's load handler can read them, and push
  // updates to the markers source as results stream in.
  const companiesRef = useRef<Company[]>(companies);
  useEffect(() => {
    companiesRef.current = companies;
    const map = mapRef.current;
    if (!map || !readyRef.current) return;
    const src = map.getSource("companies") as GeoJSONSource | undefined;
    src?.setData(toPoints(companies));
  }, [companies]);

  // Highlight the marker for the company hovered in the results list.
  const hoveredId = useHover((s) => s.hoveredId);
  const prevHoverRef = useRef<string | null>(null);
  useEffect(() => {
    const map = mapRef.current;
    if (!map || !readyRef.current) return;
    if (prevHoverRef.current) {
      map.setFeatureState({ source: "companies", id: prevHoverRef.current }, { hover: false });
    }
    if (hoveredId) {
      map.setFeatureState({ source: "companies", id: hoveredId }, { hover: true });
    }
    prevHoverRef.current = hoveredId;
  }, [hoveredId]);

  useEffect(() => {
    if (!containerRef.current || mapRef.current) return;

    const map = new MapLibreMap({
      container: containerRef.current,
      style: baseStyle,
      center: [10, 30],
      zoom: 1.4,
      attributionControl: { compact: true },
    });
    mapRef.current = map;
    map.addControl(new NavigationControl({ showCompass: false }), "bottom-right");

    const clearHover = () => {
      if (hoveredRef.current !== null) {
        map.setFeatureState(
          { source: "countries", id: hoveredRef.current },
          { hover: false },
        );
        hoveredRef.current = null;
      }
    };

    map.on("load", () => {
      map.addSource("countries", {
        type: "geojson",
        data: countries,
        generateId: true,
      });

      // Companies layer — populated as results stream in. promoteId maps each
      // feature's `key` property to its feature id so hover-highlight works.
      map.addSource("companies", {
        type: "geojson",
        data: { type: "FeatureCollection", features: [] },
        promoteId: "key",
      });
      map.addLayer({
        id: "company-points",
        type: "circle",
        source: "companies",
        paint: {
          "circle-radius": [
            "case",
            ["boolean", ["feature-state", "hover"], false],
            9,
            5,
          ],
          "circle-color": [
            "case",
            ["boolean", ["feature-state", "hover"], false],
            "#f59e0b", // amber-500 when highlighted from the list
            "#059669", // emerald-600 default
          ],
          "circle-stroke-color": "#ffffff",
          "circle-stroke-width": [
            "case",
            ["boolean", ["feature-state", "hover"], false],
            2.5,
            1,
          ],
          "circle-opacity": 0.9,
        },
      });

      const popup = new Popup({ closeButton: false, closeOnClick: true, offset: 10 });
      map.on("mouseenter", "company-points", () => (map.getCanvas().style.cursor = "pointer"));
      map.on("mouseleave", "company-points", () => {
        map.getCanvas().style.cursor = "";
        popup.remove();
      });
      map.on("click", "company-points", (e: MapLayerMouseEvent) => {
        const f = e.features?.[0];
        if (!f || f.geometry.type !== "Point") return;
        const p = (f.properties ?? {}) as Record<string, string | undefined>;

        // Build the popup as DOM with textContent — company data comes from
        // publicly editable sources (OSM), so it must never be treated as HTML.
        const root = document.createElement("div");
        root.style.font = "500 13px/1.4 system-ui";
        root.style.color = "#0f172a";
        root.style.minWidth = "160px";

        const name = document.createElement("div");
        name.style.fontWeight = "600";
        name.textContent = p.name ?? "";
        root.appendChild(name);

        const category = document.createElement("div");
        category.style.color = "#64748b";
        category.style.textTransform = "capitalize";
        category.textContent = p.category ?? "";
        root.appendChild(category);

        // Contact chips — each a validated link (tel:/mailto:/https:).
        const chips = document.createElement("div");
        chips.style.cssText = "display:flex;flex-wrap:wrap;gap:6px;margin-top:6px";
        for (const { kind, label } of CONTACT_CHIPS) {
          const href = contactHref(kind, p[kind] ?? "");
          if (!href) continue;
          const a = document.createElement("a");
          a.href = href; // property assignment; scheme validated above
          a.target = "_blank";
          a.rel = "noopener noreferrer";
          a.textContent = label;
          a.title = p[kind] ?? "";
          a.style.cssText =
            "display:inline-block;padding:2px 7px;border-radius:9999px;background:#ecfdf5;color:#047857;font-weight:600;font-size:11px;text-decoration:none";
          chips.appendChild(a);
        }
        if (chips.childElementCount > 0) root.appendChild(chips);

        popup.setLngLat(f.geometry.coordinates as [number, number]).setDOMContent(root).addTo(map);
      });

      readyRef.current = true;
      const src = map.getSource("companies") as GeoJSONSource | undefined;
      src?.setData(toPoints(companiesRef.current));

      map.addLayer({
        id: "country-fill",
        type: "fill",
        source: "countries",
        paint: {
          "fill-color": [
            "case",
            ["boolean", ["feature-state", "selected"], false],
            "#059669", // emerald-600 — selected
            ["boolean", ["feature-state", "hover"], false],
            "#34d399", // emerald-400 — hover
            "#6366f1", // indigo-500 — default
          ],
          "fill-opacity": [
            "case",
            ["boolean", ["feature-state", "selected"], false],
            0.45,
            ["boolean", ["feature-state", "hover"], false],
            0.3,
            0.08,
          ],
        },
      });

      map.addLayer({
        id: "country-line",
        type: "line",
        source: "countries",
        paint: {
          "line-color": [
            "case",
            ["boolean", ["feature-state", "selected"], false],
            "#047857",
            "#6366f1",
          ],
          "line-width": [
            "case",
            ["boolean", ["feature-state", "selected"], false],
            2.2,
            0.6,
          ],
        },
      });

      map.on("mousemove", "country-fill", (e: MapLayerMouseEvent) => {
        map.getCanvas().style.cursor = "pointer";
        const f = e.features?.[0];
        if (!f || f.id === undefined) return;
        if (hoveredRef.current === f.id) return;
        clearHover();
        hoveredRef.current = f.id;
        map.setFeatureState(
          { source: "countries", id: f.id },
          { hover: true },
        );
      });

      map.on("mouseleave", "country-fill", () => {
        map.getCanvas().style.cursor = "";
        clearHover();
      });

      map.on("click", "country-fill", (e: MapLayerMouseEvent) => {
        const f = e.features?.[0];
        if (!f || f.id === undefined) return;

        if (selectedRef.current !== null) {
          map.setFeatureState(
            { source: "countries", id: selectedRef.current },
            { selected: false },
          );
        }
        selectedRef.current = f.id;
        map.setFeatureState(
          { source: "countries", id: f.id },
          { selected: true },
        );

        const props = (f.properties ?? {}) as {
          name?: string;
          osmAreaId?: number;
        };
        const bbox = bboxOf(f as Feature);
        const next: Region = {
          name: props.name ?? "Unknown region",
          osmAreaId: props.osmAreaId ?? 0,
          bbox,
        };
        setRegionRef.current(next);

        map.fitBounds(
          [
            [bbox[0], bbox[1]],
            [bbox[2], bbox[3]],
          ],
          { padding: 48, maxZoom: 6, duration: 700 },
        );
      });
    });

    return () => {
      map.remove();
      mapRef.current = null;
    };
  }, []);

  const wrapperClass = fullScreen
    ? "absolute inset-0"
    : "relative h-[340px] w-full overflow-hidden rounded-2xl border border-slate-200 shadow-sm sm:h-[420px] lg:h-full dark:border-slate-800";

  const chipClass = fullScreen
    ? "pointer-events-none absolute left-1/2 top-4 z-10 flex -translate-x-1/2 items-center gap-2 rounded-full bg-white/90 px-3 py-1.5 text-sm font-medium text-slate-800 shadow-md ring-1 ring-slate-200 backdrop-blur dark:bg-slate-900/90 dark:text-slate-100 dark:ring-slate-700"
    : "pointer-events-none absolute left-3 top-3 z-10 flex items-center gap-2 rounded-full bg-white/90 px-3 py-1.5 text-sm font-medium text-slate-800 shadow-md ring-1 ring-slate-200 backdrop-blur dark:bg-slate-900/90 dark:text-slate-100 dark:ring-slate-700";

  return (
    <div className={wrapperClass}>
      <div ref={containerRef} className="h-full w-full" />
      <AnimatePresence>
        {region && (
          <motion.div
            key={region.name}
            initial={{ opacity: 0, y: -8, scale: 0.96 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -8, scale: 0.96 }}
            transition={{ type: "spring", stiffness: 320, damping: 26 }}
            className={chipClass}
          >
            <span className="h-2 w-2 rounded-full bg-emerald-500" />
            {region.name}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
