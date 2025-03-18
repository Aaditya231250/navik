import React from 'react';
import { View, StyleSheet } from 'react-native';
import SearchBarWithSuggestions from '../components/searchBarAutoSuggest'; // Import the component we created earlier

export default function HomeScreen() {
  return (
    <View style={styles.container}>
      <SearchBarWithSuggestions />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 16,
    backgroundColor: '#fff',
  }
});
