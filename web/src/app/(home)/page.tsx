"use client";
import { env } from "@/env";
import { useEffect, useState } from "react";
import Shell from "@/components/shell/shell";
import type { TreantChart } from "@/lib/convertToTreant";
import {
  PageHeader,
  PageHeaderHeading,
  PageHeaderDescription,
} from "@/components/page-header";
import TreantDiagram from "@/components/TreantDiagram";
import { convertToTreant } from "@/lib/convertToTreant";

import { toast } from "@/hooks/use-toast";

export default function Home() {
  const [treeData, setTreeData] = useState<TreantChart | null>(null);
  const [loading, setLoading] = useState(false);
  const [target, setTarget] = useState("");
  const [maxPaths, setMaxPaths] = useState(1);
  const [searchType, setSearchType] = useState("shortest/dfs");
  const [liveUpdate, setLiveUpdate] = useState(false);
  const [delay, setDelay] = useState(100);
  const [stats, setStats] = useState({
    nodeCount: 0,
    completePaths: 0,
    executionTimeMs: 0,
  });

  useEffect(() => {
    fetch(`${env.NEXT_PUBLIC_BACKEND_HTTP}/api/config`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ elementPath: "configs/elements.json" }),
    })
      .then((res) => {
        if (!res.ok) {
          toast({
            title: "Error",
            description: "Something went wrong when loading elements",
          });
          return;
        }
        console.log("Config loaded");
        toast({
          title: "Elements loaded",
          description: "You can now start generating trees",
        });
      })
      .catch(console.error);
  }, []);

  const loadTreeAPI = async () => {
    try {
      const res = await fetch(
        `${env.NEXT_PUBLIC_BACKEND_HTTP}/api/${searchType}`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            target,
            maxPaths,
            delay: 0,
          }),
        }
      );

      if (!res.ok) {
        const errorText = await res.text();
        toast({
          title: "Error",
          description: errorText || "Failed to generate tree",
        });
        setLoading(false);
        return;
      }

      const data = await res.json();
      const converted = convertToTreant(data.tree);
      setTreeData(converted);

      if (data.nodeCount !== undefined) {
        setStats((prev) => ({ ...prev, nodeCount: data.nodeCount }));
      }

      if (data.completePaths !== undefined) {
        setStats((prev) => ({ ...prev, completePaths: data.completePaths }));
      }

      if (data.executionTimeMs !== undefined) {
        setStats((prev) => ({
          ...prev,
          executionTimeMs: data.executionTimeMs,
        }));
      }

      toast({
        title: "Success",
        description: "Tree generated without live update.",
      });
    } catch (err) {
      console.error("API request error:", err);
      toast({
        title: "Error",
        description: "Something went wrong with the API request.",
      });
    } finally {
      setLoading(false);
    }
  };

  const loadTreeWS = async () => {
    try {
      const ws = new WebSocket(
        `${env.NEXT_PUBLIC_BACKEND_WS}/ws/${searchType}`
      );

      ws.onopen = () => {
        console.log("WebSocket connection opened");
        ws.send(
          JSON.stringify({
            target,
            maxPaths,
            delay,
          })
        );
      };

      ws.onmessage = (event) => {
        let msg;

        try {
          msg = JSON.parse(event.data);
        } catch {
          console.error("Invalid JSON received:", event.data);
          toast({
            title: "Error",
            description: "Received invalid data from the server.",
          });
          ws.close();
          setLoading(false);
          return;
        }

        console.log("Received:", msg);

        if (msg.error) {
          toast({
            title: "Error",
            description: msg.error,
          });
        }

        if (msg.nodeCount !== undefined) {
          setStats((prev) => ({ ...prev, nodeCount: msg.nodeCount }));
        }

        if (msg.completePaths !== undefined) {
          setStats((prev) => ({ ...prev, completePaths: msg.completePaths }));
        }

        if (msg.executionTimeMs !== undefined) {
          setStats((prev) => ({
            ...prev,
            executionTimeMs: msg.executionTimeMs,
          }));
        }

        if (msg.tree) {
          const converted = convertToTreant(msg.tree);
          setTreeData(converted);
          console.log("converted");
        }

        if (msg.done) {
          setLoading(false);
          toast({
            title: "Success",
            description:
              "Element recipes tree has been generated successfully.",
          });
          ws.close();
        }
      };

      ws.onerror = (e) => {
        console.error("WebSocket error:", e);
        toast({
          title: "Error",
          description: "WebSocket connection failed.",
        });
        setTreeData(null);
        setLoading(false);
      };

      ws.onclose = () => {
        setLoading(false);
        console.log("WebSocket closed");
      };
    } catch (err) {
      console.error("WebSocket setup error", err);
      setTreeData(null);
      setLoading(false);
    }
  };

  const handleGenerate = async () => {
    setLoading(true);
    setTreeData(null);
    if (liveUpdate) {
      loadTreeWS();
    } else {
      loadTreeAPI();
    }
  };

  return (
    <Shell className="md:pb-10">
      <PageHeader>
        <PageHeaderHeading>Tubes Stima</PageHeaderHeading>
        <PageHeaderDescription>
          Element Recipe Tree Generation
        </PageHeaderDescription>
      </PageHeader>

      <div className="mt-6 space-y-4">
        <div className="flex gap-3 items-center flex-wrap">
          <select
            value={searchType}
            onChange={(e) => setSearchType(e.target.value)}
            className="border px-3 py-2 rounded text-sm"
            disabled={loading}
          >
            <option value="shortest/dfs">DFS</option>
            <option value="shortest/bfs">BFS</option>
            <option value="dfs">DFS Multi-Recipe</option>
            <option value="bfs">BFS Multi-Recipe</option>
          </select>
          <input
            type="text"
            value={target}
            onChange={(e) => setTarget(e.target.value)}
            className="border px-3 py-2 rounded text-sm w-64"
            placeholder="Enter element name"
            disabled={loading}
          />
          {(searchType === "bfs" || searchType === "dfs") && (
            <input
              type="number"
              value={maxPaths}
              onChange={(e) => setMaxPaths(Number(e.target.value))}
              className="border px-3 py-2 rounded text-sm w-24"
              placeholder="Max paths"
              min={1}
              disabled={loading}
            />
          )}

          <button
            onClick={handleGenerate}
            className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 text-sm disabled:opacity-50 disabled:hover:bg-blue-600"
            disabled={loading}
          >
            Generate Tree
          </button>
        </div>

        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={liveUpdate}
              onChange={(e) => setLiveUpdate(e.target.checked)}
              disabled={loading}
            />
            Show live update
          </label>

          {liveUpdate && (
            <input
              type="number"
              value={delay}
              onChange={(e) => setDelay(Number(e.target.value))}
              className="border px-3 py-2 rounded text-sm w-24"
              placeholder="Delay (ms)"
              min={100}
              disabled={loading}
            />
          )}
        </div>

        {treeData && (
          <div className="mt-4 text-sm text-gray-600">
            <p>
              🧠 Nodes Explored: <strong>{stats.nodeCount}</strong>
            </p>
            <p>
              🌱 Complete Paths: <strong>{stats.completePaths}</strong>
            </p>
            <p>
              ⏱️ Execution Time: <strong>{stats.executionTimeMs} ms</strong>
            </p>
          </div>
        )}

        {treeData ? (
          <TreantDiagram data={treeData} />
        ) : loading ? (
          <p className="text-sm text-gray-500">Loading tree...</p>
        ) : (
          <p className="text-sm text-blue-600">Start generating trees...</p>
        )}
      </div>
    </Shell>
  );
}
