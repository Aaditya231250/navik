import React from "react";
import { Stack } from "expo-router";

export default function RideJourneyLayout() {
  return (
    <Stack screenOptions={{ headerShown: false }}>
      <Stack.Screen name="index" />
      <Stack.Screen name="driver-assigned" />
      <Stack.Screen name="in-progress" />
      <Stack.Screen name="completed" />
    </Stack>
  );
}
