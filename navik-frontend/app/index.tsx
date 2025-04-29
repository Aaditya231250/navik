import { View, Text, Image, StyleSheet, StatusBar, Dimensions } from "react-native";
import React, { useEffect, useState } from "react";
import { useRouter } from "expo-router";
import AsyncStorage from "@react-native-async-storage/async-storage";

export default function Index() {
  const router = useRouter();
  const { width } = Dimensions.get("window");
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Check if user is authenticated
    const checkAuthStatus = async () => {
      try {
        // Try to get the auth token from storage
        const userToken = await AsyncStorage.getItem('userToken');
        
        // Wait for splash screen minimum time
        await new Promise(resolve => setTimeout(resolve, 2000));
        
        // Navigate based on authentication status
        if (userToken) {
          // User is authenticated, go to home
          router.replace("/home");
        } else {
          // User is not authenticated, go to login
          router.replace("/auth/login");
        }
      } catch (error) {
        console.error("Failed to check auth status:", error);
        // On error, default to login page
        router.replace("/auth/login");
      } finally {
        setIsLoading(false);
      }
    };

    checkAuthStatus();
  }, [router]);

  return (
    <>
      <StatusBar barStyle="light-content" backgroundColor="#000" />
      <View style={styles.container}>
        {/* Logo */}
        {/* <Image
          source={require("../assets/images/react-logo.png")}
          style={[styles.logo, { width: width * 0.3, height: width * 0.3 }]}
        /> */}
        
        {/* Title */}
        <Text style={styles.title}>NAVIK</Text>
      </View>
    </>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#000", // Dark background
    justifyContent: "center", // Center content vertically
    alignItems: "center", // Center content horizontally
  },
  logo: {
    resizeMode: "contain",
    marginBottom: 20,
  },
  title: {
    fontSize: 36,
    fontWeight: "bold",
    color: "#fff", // White text for contrast
    letterSpacing: 2,
    marginTop: 20,
  },
});
