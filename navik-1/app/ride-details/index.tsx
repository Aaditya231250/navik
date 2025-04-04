import React, { useState } from "react";
import {
  View,
  Text,
  TouchableOpacity,
  Image,
  FlatList,
  SafeAreaView,
} from "react-native";
import { useRouter } from "expo-router";
import { MaterialIcons } from "@expo/vector-icons";
import MapView, { Marker } from "react-native-maps";

export default function RideDetailsScreen() {
  const router = useRouter();
  const [selectedRide, setSelectedRide] = useState(null);

  const rideOptions = [
    {
      id: "uber-go",
      name: "Uber Go",
      price: "₹170.71",
      time: "8:46pm - 4 min away",
      image: require("@/assets/images/uber-go.png"),
      isFaster: true,
    },
    {
      id: "auto",
      name: "Auto",
      price: "₹170.71",
      time: "8:46pm - 4 min away",
      image: require("@/assets/images/auto.png"),
      oldPrice: "₹189.71",
    },
    {
      id: "uber-premier",
      name: "Uber Premier",
      price: "₹223.63",
      time: "8:46pm - 5 min away",
      image: require("@/assets/images/uber-premier.png"),
    },
  ];

  const renderRideOption = ({ item }) => (
    <TouchableOpacity
      onPress={() => setSelectedRide(item.id)}
      className={`flex-row items-center justify-between p-4 rounded-lg border ${
        selectedRide === item.id ? "border-black" : "border-gray-200"
      }`}
    >
      <View className="flex-row items-center">
        <Image source={item.image} className="w-12 h-12 mr-4" />
        <View>
          <Text className="text-lg font-semibold">{item.name}</Text>
          <Text className="text-sm text-gray-500">{item.time}</Text>
          {item.isFaster && (
            <Text className="text-xs text-blue-500 font-bold mt-1">Faster</Text>
          )}
        </View>
      </View>
      <View className="items-end">
        <Text className="text-lg font-bold">{item.price}</Text>
        {item.oldPrice && (
          <Text className="text-sm text-gray-400 line-through">
            {item.oldPrice}
          </Text>
        )}
      </View>
    </TouchableOpacity>
  );

  return (
    <View className="flex-1">
      {/* Full Screen Map */}
      <MapView
        style={{ flex: 1 }}
        initialRegion={{
          latitude: 26.4753,
          longitude: 73.1149,
          latitudeDelta: 0.01,
          longitudeDelta: 0.01,
        }}
      >
        <Marker
          coordinate={{ latitude: 26.4753, longitude: 73.1149 }}
          title="IIT Jodhpur"
          description="Indian Institute of Technology, Jodhpur"
        />
      </MapView>

      {/* Back Button */}
      <TouchableOpacity
        onPress={() => router.back()}
        className="absolute top-6 left-4 z-10 bg-white p-2 rounded-full shadow"
      >
        <MaterialIcons name="arrow-back" size={24} color="#000" />
      </TouchableOpacity>

      {/* Ride Options Overlay */}
      <View className="absolute bottom-0 left-0 right-0 bg-white rounded-t-2xl p-4 shadow-lg">
        <Text className="text-xl font-bold mb-4">Choose a trip</Text>

        <FlatList
          data={rideOptions}
          renderItem={renderRideOption}
          keyExtractor={(item) => item.id}
          ItemSeparatorComponent={() => <View style={{ height: 10 }} />}
        />

        {/* Payment Option */}
        <TouchableOpacity
          onPress={() => {}}
          className="flex-row items-center justify-between mt-6 px-4 py-3 bg-gray-100 rounded-lg"
        >
          <MaterialIcons name="payment" size={24} color="#000" />
          <Text className="text-gray-700 flex-grow ml-2">
            user.name@okhdfcbank
          </Text>
          <MaterialIcons name="keyboard-arrow-right" size={24} color="#000" />
        </TouchableOpacity>

        {/* Confirm Button */}
        <TouchableOpacity
          onPress={() => alert(`You selected ${selectedRide}`)}
          disabled={!selectedRide}
          className={`mt-4 px-6 py-3 rounded-lg ${
            selectedRide ? "bg-black" : "bg-gray-300"
          }`}
        >
          <Text className="text-white text-center text-lg font-semibold">
            {selectedRide
              ? `Choose ${
                  rideOptions.find((r) => r.id === selectedRide)?.name
                }`
              : "Select a Ride"}
          </Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}
