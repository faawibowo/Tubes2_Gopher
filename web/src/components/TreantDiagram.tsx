"use client";

import { useEffect, useRef } from "react";
import type { TreantChart } from "@/lib/convertToTreant";
import Panzoom from "@panzoom/panzoom";

interface Props {
  data: TreantChart;
}

export default function TreantDiagram({ data }: Props) {
  const wrapperRef = useRef<HTMLDivElement>(null);
  const treeRef = useRef<HTMLDivElement>(null);
  const panzoomRef = useRef<ReturnType<typeof Panzoom> | null>(null);

  useEffect(() => {
    if (!treeRef.current) return;

    const panzoom = Panzoom(treeRef.current!, {
      minScale: 0.1,
      maxScale: 5,
      panOnly: true,
    });
    panzoomRef.current = panzoom;

    wrapperRef.current?.addEventListener("wheel", panzoom.zoomWithWheel);

    return () => {
      panzoom.destroy();
    };
  }, []);

  useEffect(() => {
    if (typeof window === "undefined" || !data) return;

    const script = document.createElement("script");
    script.src = "/treant/Treant.min.js";
    script.defer = true;
    script.onload = () => {
      new window.Treant(data);
    };

    document.body.appendChild(script);

    return () => {
      document.body.removeChild(script);
    };
  }, [data]);

  return (
    <div
      ref={wrapperRef}
      style={{
        width: "100%",
        height: "100%",
        maxHeight: "600px", // or any height that fits your design
        border: "1px solid #ccc",
        position: "relative",
        overflow: "hidden", // ✅ clip internal overflows
        backgroundColor: "#fff",
      }}
    >
      <div
        ref={treeRef}
        style={{
          width: "900px",
          height: "900px",
          position: "absolute",
          touchAction: "none",
          cursor: "grab",
        }}
      >
        <div
          id="tree"
          style={{
            width: "fit-content",
            height: "fit-content",
            position: "absolute",
            top: "50%",
            left: "50%",
            transform: "translate(-50%, -50%)",
            padding: "2rem",
            cursor: "grab",
          }}
        />
      </div>
    </div>
  );
}
