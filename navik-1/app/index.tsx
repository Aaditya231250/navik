// app/index.tsx
import { View, Text, Image } from "react-native";
import { useRouter } from "expo-router";
import { Pressable } from "react-native";
import "../global.css"

export default function Index() {
  const router = useRouter();

  return (
    <View className="flex-1 justify-center items-center bg-slate-100">
      {/* Logo Section */}
      <View className="items-center mb-8">
        <Image
          source={require("../assets/images/react-logo.png")} // Add your logo in assets
          className="w-32 h-32 mb-4 rounded-lg"
        />
        <Text className="text-3xl font-bold text-blue-600">Welcome to Navik</Text>
        <Text className="text-gray-600 mt-2">Your Navigation Assistant</Text>
      </View>

      {/* Action Button */}
      <Pressable
        onPress={() => router.push("/home")}
        className="bg-blue-500 px-6 py-3 rounded-lg shadow-lg active:bg-blue-600"
      >
        <Text className="text-white text-lg font-semibold">Get Started</Text>
      </Pressable>

      {/* Footer */}
      <Text className="absolute bottom-8 text-gray-500 text-sm">
        v1.0.0 • navik.com
      </Text>
    </View>
  );
}
