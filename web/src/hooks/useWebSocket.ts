import { useEffect, useState, useCallback, useRef } from "react";
import { HealthResponse, StatusResponse, WebSocketMessage } from "../types/api";

interface UseWebSocketReturn {
  health: HealthResponse | null;
  status: StatusResponse | null;
  loading: boolean;
  error: string | null;
  wsConnected: boolean;
}

export function useWebSocket(): UseWebSocketReturn {
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [status, setStatus] = useState<StatusResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [wsConnected, setWsConnected] = useState(false);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const isConnectingRef = useRef(false);

  const connectWebSocket = useCallback(() => {
    // Prevent duplicate connections
    if (
      isConnectingRef.current ||
      (wsRef.current && wsRef.current.readyState === WebSocket.OPEN)
    ) {
      console.log("Already connecting or connected, skipping...");
      return;
    }

    isConnectingRef.current = true;

    // Clear any existing reconnect timeout
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    // Close existing connection if any
    if (wsRef.current) {
      wsRef.current.close();
    }

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/api/ws`;

    console.log("Connecting to WebSocket:", wsUrl);

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log("WebSocket connected");
      setWsConnected(true);
      setError(null);
      setLoading(false);
      reconnectAttemptsRef.current = 0;
      isConnectingRef.current = false;
    };

    ws.onmessage = (event) => {
      try {
        const data: WebSocketMessage = JSON.parse(event.data);

        if (data.type === "status_update") {
          setHealth(data.health);
          setStatus(data.status);
          setError(null);
        }
      } catch (err) {
        console.error("Failed to parse WebSocket message:", err);
        setError("Failed to parse server data");
      }
    };

    ws.onerror = (event) => {
      console.error("WebSocket error:", event);
      setError("WebSocket connection error");
      setWsConnected(false);
      isConnectingRef.current = false;
    };

    ws.onclose = (event) => {
      console.log("WebSocket closed:", event.code, event.reason);
      setWsConnected(false);
      wsRef.current = null;
      isConnectingRef.current = false;

      // Attempt to reconnect with exponential backoff
      reconnectAttemptsRef.current += 1;
      const delay = Math.min(
        1000 * Math.pow(2, reconnectAttemptsRef.current),
        30000,
      );

      console.log(
        `Reconnecting in ${delay}ms (attempt ${reconnectAttemptsRef.current})`,
      );

      reconnectTimeoutRef.current = setTimeout(() => {
        connectWebSocket();
      }, delay);
    };
  }, []);

  useEffect(() => {
    connectWebSocket();

    // Cleanup on unmount
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return {
    health,
    status,
    loading,
    error,
    wsConnected,
  };
}
