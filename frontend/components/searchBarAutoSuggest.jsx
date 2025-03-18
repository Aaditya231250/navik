import React, { useState, useEffect, useRef } from 'react';
import {
  View,
  Text,
  TextInput,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  Animated,
  Alert,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';

const SearchBarWithSuggestions = () => {
  const [searchQuery, setSearchQuery] = useState('');
  const [suggestions, setSuggestions] = useState([]);
  const [isConnected, setIsConnected] = useState(false);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [selectedPlace, setSelectedPlace] = useState(null);
  const ws = useRef(null);
  const suggestionHeight = useRef(new Animated.Value(0)).current;

  // Connect to WebSocket
  useEffect(() => {
    connectWebSocket();
    
    return () => {
      if (ws.current && ws.current.readyState === WebSocket.OPEN) {
        ws.current.close();
      }
    };
  }, []);

  // Animation for suggestions panel
  useEffect(() => {
    Animated.timing(suggestionHeight, {
      toValue: showSuggestions ? 200 : 0,
      duration: 300,
      useNativeDriver: false,
    }).start();
  }, [showSuggestions]);

  const connectWebSocket = () => {
    // Replace with your Go server WebSocket URL
    ws.current = new WebSocket('ws://localhost:8080/search');
    
    ws.current.onopen = () => {
      console.log('WebSocket connection established');
      setIsConnected(true);
    };
    
    ws.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      // Check if this is a place details response
      if (data.placeDetails) {
        // Handle place details response
        console.log('Received place details:', data.placeDetails);
        // You can update your UI or navigate to a details screen here
        Alert.alert('Place Details Received', 
          `Details for ${data.placeDetails.name} have been received.`);
      } else {
        // This is a search suggestions response
        setSuggestions(data.places || []);
        setShowSuggestions(data.places && data.places.length > 0);
      }
    };
    
    ws.current.onerror = (error) => {
      console.error('WebSocket error:', error);
      setIsConnected(false);
    };
    
    ws.current.onclose = () => {
      console.log('WebSocket connection closed');
      setIsConnected(false);
      // Attempt to reconnect after a delay
      setTimeout(connectWebSocket, 3000);
    };
  };

  const handleSearch = (text) => {
    setSearchQuery(text);
    
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      // Send search query to server on each keystroke
      ws.current.send(JSON.stringify({ 
        type: 'search',
        query: text 
      }));
      
      // Show suggestions panel if text is not empty
      setShowSuggestions(text.length > 0);
    }
  };

  const selectSuggestion = (item) => {
    setSearchQuery(item.name);
    setSelectedPlace(item);
    setShowSuggestions(false);
    
    // Send selected place ID back to server for further processing
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      console.log(`Sending selected place ID: ${item.id} to server`);
      ws.current.send(JSON.stringify({
        type: 'placeSelected',
        placeId: item.id
      }));
    }
  };

  const clearSearch = () => {
    setSearchQuery('');
    setShowSuggestions(false);
    setSelectedPlace(null);
  };

  return (
    <View style={styles.container}>
      <View style={styles.searchBarContainer}>
        <View style={[styles.searchBar, isConnected ? styles.connected : styles.disconnected]}>
          <Ionicons name="search" size={20} color="#666" style={styles.searchIcon} />
          <TextInput
            style={styles.input}
            placeholder="Search places..."
            value={searchQuery}
            onChangeText={handleSearch}
            onFocus={() => searchQuery.length > 0 && setShowSuggestions(true)}
          />
          {searchQuery.length > 0 && (
            <TouchableOpacity onPress={clearSearch} style={styles.clearButton}>
              <Ionicons name="close-circle" size={20} color="#666" />
            </TouchableOpacity>
          )}
        </View>
        {isConnected ? (
          <View style={styles.statusIndicator}>
            <View style={styles.connectedDot} />
          </View>
        ) : (
          <TouchableOpacity onPress={connectWebSocket} style={styles.reconnectButton}>
            <Ionicons name="refresh" size={20} color="#666" />
          </TouchableOpacity>
        )}
      </View>
      
      {selectedPlace && (
        <View style={styles.selectedPlaceContainer}>
          <Text style={styles.selectedPlaceTitle}>Selected Place:</Text>
          <Text style={styles.selectedPlaceName}>{selectedPlace.name}</Text>
          <Text style={styles.selectedPlaceAddress}>{selectedPlace.address}</Text>
        </View>
      )}
      
      <Animated.View style={[styles.suggestionsContainer, { height: suggestionHeight }]}>
        <FlatList
          data={suggestions}
          keyExtractor={(item) => item.id.toString()}
          renderItem={({ item }) => (
            <TouchableOpacity
              style={styles.suggestionItem}
              onPress={() => selectSuggestion(item)}
            >
              <Ionicons name="location" size={18} color="#5751D9" style={styles.locationIcon} />
              <View>
                <Text style={styles.suggestionText}>{item.name}</Text>
                <Text style={styles.suggestionSubtext}>{item.address}</Text>
              </View>
            </TouchableOpacity>
          )}
          ListEmptyComponent={
            searchQuery.length > 0 ? (
              <Text style={styles.noResults}>No places found</Text>
            ) : null
          }
        />
      </Animated.View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    padding: 16,
    position: 'relative',
    zIndex: 1,
  },
  searchBarContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  searchBar: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    height: 50,
    borderRadius: 25,
    paddingHorizontal: 15,
    backgroundColor: '#f5f5f5',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 2,
  },
  connected: {
    borderColor: '#5751D9',
    borderWidth: 1,
  },
  disconnected: {
    borderColor: '#ff6b6b',
    borderWidth: 1,
  },
  searchIcon: {
    marginRight: 10,
  },
  input: {
    flex: 1,
    fontSize: 16,
  },
  clearButton: {
    padding: 5,
  },
  statusIndicator: {
    marginLeft: 10,
  },
  connectedDot: {
    width: 10,
    height: 10,
    borderRadius: 5,
    backgroundColor: '#5751D9',
  },
  reconnectButton: {
    marginLeft: 10,
    padding: 5,
  },
  suggestionsContainer: {
    backgroundColor: 'white',
    borderRadius: 10,
    marginTop: 5,
    overflow: 'hidden',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 3 },
    shadowOpacity: 0.2,
    shadowRadius: 5,
    elevation: 5,
  },
  suggestionItem: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 15,
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
  },
  locationIcon: {
    marginRight: 10,
  },
  suggestionText: {
    fontSize: 16,
    fontWeight: '500',
  },
  suggestionSubtext: {
    fontSize: 14,
    color: '#666',
    marginTop: 2,
  },
  noResults: {
    padding: 15,
    textAlign: 'center',
    color: '#666',
  },
  selectedPlaceContainer: {
    marginTop: 15,
    padding: 15,
    backgroundColor: '#f0f8ff',
    borderRadius: 10,
    borderLeftWidth: 4,
    borderLeftColor: '#5751D9',
  },
  selectedPlaceTitle: {
    fontSize: 14,
    color: '#666',
    marginBottom: 5,
  },
  selectedPlaceName: {
    fontSize: 16,
    fontWeight: 'bold',
  },
  selectedPlaceAddress: {
    fontSize: 14,
    color: '#333',
    marginTop: 2,
  },
});

export default SearchBarWithSuggestions;
