"use client";

import { useEffect, useRef } from "react";

interface Props {
  data: any;
}

export default function TreantDiagram({ data }: Props) {
  const treeRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (typeof window === "undefined" || !data) return;

    const script = document.createElement("script");
    script.src = "/treant/Treant.min.js";
    script.defer = true;
    script.onload = () => {
      new window.Treant(data); // safe to run after script load
    };

    document.body.appendChild(script);

    return () => {
      document.body.removeChild(script);
    };
  }, [data]);

  return (
    <div
      id="tree"
      ref={treeRef}
      style={{
        width: "100%",
        height: "auto",
        overflowX: "auto",
        overflowY: "auto",
        padding: "2rem",
      }}
    />
  );
}
