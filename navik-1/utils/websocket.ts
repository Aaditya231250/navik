// app/utils/websocket.ts
import { useEffect, useRef, useState, useCallback } from 'react';

export interface SearchPlace {
  id: string;
  name: string;
  address: string;
}

export interface PlaceDetails {
  id: string;
  name: string;
  address: string;
  latitude: number;
  longitude: number;
  phoneNumber?: string;
  website?: string;
  rating?: number;
}

// Define the messages our WebSocket will handle
type WebSocketMessage = 
  | { type: 'search', query: string }
  | { type: 'placeSelected', placeId: string };

// Define response types
interface SearchResponse {
  places: SearchPlace[];
}

interface PlaceDetailsResponse {
  placeDetails: PlaceDetails;
}

export const useMapWebSocket = (wsUrl: string = 'ws://localhost:8082/search') => {
  const [isConnected, setIsConnected] = useState(false);
  const [searchResults, setSearchResults] = useState<SearchPlace[]>([]);
  const [selectedPlace, setSelectedPlace] = useState<PlaceDetails | null>(null);
  const [error, setError] = useState<string | null>(null);
  
  const socket = useRef<WebSocket | null>(null);
  const connectingRef = useRef(false);

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (socket.current?.readyState === WebSocket.OPEN || connectingRef.current) {
      return;
    }

    connectingRef.current = true;
    try {
      console.log(`Connecting to WebSocket at ${wsUrl}`);
      const ws = new WebSocket(wsUrl);
      
      ws.onopen = () => {
        console.log('WebSocket connection established');
        setIsConnected(true);
        setError(null);
        connectingRef.current = false;
      };
      
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          
          // Handle search results response
          if (data.places) {
            console.log(`Received ${data.places.length} search results`);
            setSearchResults(data.places);
          } 
          // Handle place details response
          else if (data.placeDetails) {
            console.log(`Received details for place: ${data.placeDetails.name}`);
            setSelectedPlace(data.placeDetails);
          }
        } catch (err) {
          console.error('Error parsing WebSocket message:', err);
          setError('Failed to parse server response');
        }
      };
      
      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        setError('WebSocket connection error');
        setIsConnected(false);
        connectingRef.current = false;
      };
      
      ws.onclose = () => {
        console.log('WebSocket connection closed');
        setIsConnected(false);
        connectingRef.current = false;
      };
      
      socket.current = ws;
    } catch (err) {
      console.error('Failed to create WebSocket connection:', err);
      setError('Failed to connect to search service');
      connectingRef.current = false;
    }
  }, [wsUrl]);

  // Search for places
  const searchPlaces = useCallback((query: string) => {
    if (!socket.current || socket.current.readyState !== WebSocket.OPEN) {
      setError('WebSocket not connected');
      connect();
      return;
    }
    
    console.log(`Searching for: ${query}`);
    const message: WebSocketMessage = { type: 'search', query };
    socket.current.send(JSON.stringify(message));
  }, [connect]);

  // Get place details
  const selectPlace = useCallback((placeId: string) => {
    if (!socket.current || socket.current.readyState !== WebSocket.OPEN) {
      setError('WebSocket not connected');
      connect();
      return;
    }
    
    console.log(`Selecting place ID: ${placeId}`);
    const message: WebSocketMessage = { type: 'placeSelected', placeId };
    socket.current.send(JSON.stringify(message));
  }, [connect]);

  // Connect on component mount
  useEffect(() => {
    connect();
    
    // Clean up on unmount
    return () => {
      if (socket.current) {
        socket.current.close();
        socket.current = null;
      }
    };
  }, [connect]);

  return {
    isConnected,
    searchPlaces,
    selectPlace,
    searchResults,
    selectedPlace,
    error,
    reconnect: connect,
  };
};
