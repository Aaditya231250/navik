import { Stack } from "expo-router";
import { useEffect, useState } from "react";
import * as SplashScreen from "expo-splash-screen";
import { View } from "react-native";
import { GestureHandlerRootView } from "react-native-gesture-handler";

export default function RootLayout() {
  const [appReady, setAppReady] = useState(false);

  useEffect(() => {
    async function prepare() {
      try {
        await SplashScreen.preventAutoHideAsync();
        // Load fonts/assets here if needed
        await new Promise(resolve => setTimeout(resolve, 2000));
      } finally {
        setAppReady(true);
        await SplashScreen.hideAsync();
      }
    }
    prepare();
  }, []);

  if (!appReady) {
    return <View style={{ flex: 1, backgroundColor: "#fff" }} />;
  }

  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
    <Stack screenOptions={{ headerShown: false }}>
      <Stack.Screen name="index" />
      <Stack.Screen name="home" />
      <Stack.Screen name="plan-ride" />
      <Stack.Screen name="ride-details" />
      <Stack.Screen name="home/profile" />
    </Stack>
    </GestureHandlerRootView>
  );
}