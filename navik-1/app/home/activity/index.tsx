// app/home/activity/index.tsx
import { View, Text } from "react-native";

export default function ActivityScreen() {
  return (
    <View style={{ flex: 1, justifyContent: "center", alignItems: "center", backgroundColor: "#fff" }}>
      <Text style={{ fontSize: 24, fontWeight: "bold" }}>Activity Screen</Text>
      <Text style={{ marginTop: 8, color: "#666" }}>Your recent activities</Text>
    </View>
  );
}
