// app/home/services/index.tsx
import { View, Text } from "react-native";

export default function ServicesScreen() {
  return (
    <View style={{ flex: 1, justifyContent: "center", alignItems: "center", backgroundColor: "#fff" }}>
      <Text style={{ fontSize: 24, fontWeight: "bold" }}>Services Screen</Text>
      <Text style={{ marginTop: 8, color: "#666" }}>Explore our services</Text>
    </View>
  );
}
