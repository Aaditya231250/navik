import { View, Text, Image, StyleSheet, StatusBar, Dimensions } from "react-native";
import React, { useEffect } from "react";
import { useRouter } from "expo-router";

export default function Index() {
  const router = useRouter();
  const { width } = Dimensions.get("window");

  useEffect(() => {
    // Navigate to the next screen after a 2-second delay
    const timeout = setTimeout(() => {
      router.replace("/home");
    }, 2000);

    // Cleanup the timeout on component unmount
    return () => clearTimeout(timeout);
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
