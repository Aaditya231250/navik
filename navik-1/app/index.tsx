import { View, Text, Image, Pressable, StyleSheet, Dimensions, ImageBackground, StatusBar, Platform } from "react-native";
import React from "react";
import { useRouter } from "expo-router";

export default function Index() {
  const router = useRouter();
  const { width, height } = Dimensions.get("window");

  return (
    <>
      <StatusBar barStyle="dark-content" backgroundColor="transparent" translucent />
      <View style={styles.container}>
        {/* Background Design Elements */}
        <View style={[styles.backgroundCircle, { top: -height * 0.2, right: -width * 0.3 }]} />
        <View style={[styles.backgroundCircle, { bottom: -height * 0.15, left: -width * 0.3 }]} />
        
        {/* Main Content */}
        <View style={styles.contentContainer}>
          {/* Logo Section */}
          <View style={styles.logoContainer}>
            <View style={styles.logoWrapper}>
              <Image
                source={require("../assets/images/react-logo.png")}
                style={[styles.logo, { width: width * 0.26, height: width * 0.26 }]}
              />
            </View>
            <Text style={styles.title}>NAVIK</Text>
            <Text style={styles.subtitle}>Your Ride-Sharing Companion</Text>
          </View>

          {/* Tagline */}
          <View style={styles.taglineContainer}>
            <Text style={styles.tagline}>Journey together with comfort and safety</Text>
          </View>

          {/* Action Buttons */}
          <View style={styles.buttonContainer}>
            <Pressable
              onPress={() => router.push("/home")}
              style={({ pressed }) => [
                styles.primaryButton,
                pressed && styles.buttonPressed,
              ]}
            >
              <Text style={styles.primaryButtonText}>Get Started</Text>
            </Pressable>
            
            {/* <Pressable
              onPress={() => router.push("/login")}
              style={({ pressed }) => [
                styles.secondaryButton,
                pressed && styles.secondaryButtonPressed,
              ]}
            >
              <Text style={styles.secondaryButtonText}>Sign In</Text>
            </Pressable> */}
          </View>

          {/* Footer */}
          <View style={styles.footerContainer}>
            <Text style={styles.footer}>v1.0.0 • navik.com</Text>
          </View>
        </View>
      </View>
    </>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f8fafc",
    position: "relative",
    overflow: "hidden",
  },
  backgroundCircle: {
    position: "absolute",
    width: Dimensions.get("window").width * 0.7,
    height: Dimensions.get("window").width * 0.7,
    borderRadius: Dimensions.get("window").width * 0.35,
    backgroundColor: "rgba(37, 99, 235, 0.08)",
  },
  contentContainer: {
    flex: 1,
    justifyContent: "space-between",
    alignItems: "center",
    paddingTop: Platform.OS === "ios" ? 60 : 50,
    paddingBottom: 30,
    paddingHorizontal: 24,
  },
  logoContainer: {
    alignItems: "center",
    marginTop: 20,
  },
  logoWrapper: {
    padding: 15,
    borderRadius: 26,
    backgroundColor: "white",
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: 10,
    elevation: 8,
    marginBottom: 24,
  },
  logo: {
    resizeMode: "contain",
    borderRadius: 15,
  },
  title: {
    fontSize: 36,
    fontWeight: "bold",
    color: "#1e40af",
    letterSpacing: 2,
    textAlign: "center",
  },
  subtitle: {
    fontSize: 16,
    color: "#64748b",
    marginTop: 8,
    textAlign: "center",
    letterSpacing: 0.5,
  },
  taglineContainer: {
    marginVertical: 40,
    paddingHorizontal: 20,
  },
  tagline: {
    fontSize: 18,
    textAlign: "center",
    color: "#475569",
    lineHeight: 26,
    fontWeight: "400",
  },
  buttonContainer: {
    width: "100%",
    marginBottom: 30,
    gap: 16,
  },
  primaryButton: {
    backgroundColor: "#2563eb",
    paddingVertical: 16,
    borderRadius: 16,
    shadowColor: "#2563eb",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 5,
    alignItems: "center",
  },
  buttonPressed: {
    backgroundColor: "#1d4ed8",
    transform: [{ scale: 0.98 }],
  },
  primaryButtonText: {
    color: "#fff",
    fontSize: 18,
    fontWeight: "600",
  },
  secondaryButton: {
    backgroundColor: "transparent",
    paddingVertical: 16,
    borderRadius: 16,
    borderWidth: 1,
    borderColor: "#cbd5e1",
    alignItems: "center",
  },
  secondaryButtonPressed: {
    backgroundColor: "#f1f5f9",
    transform: [{ scale: 0.98 }],
  },
  secondaryButtonText: {
    color: "#475569",
    fontSize: 18,
    fontWeight: "500",
  },
  footerContainer: {
    width: "100%",
    alignItems: "center",
  },
  footer: {
    fontSize: 14,
    color: "#94a3b8",
    textAlign: "center",
  },
});