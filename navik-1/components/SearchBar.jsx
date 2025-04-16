import { View, TextInput, TouchableOpacity, Platform } from "react-native";
import React from "react";
import { MaterialIcons } from "@expo/vector-icons";
import { useRouter } from "expo-router";

const SearchBar = () => {
  const router = useRouter();

  return (
    <TouchableOpacity
      onPress={() => router.push("/plan-ride")}
      className="bg-white h-[7.5vh] rounded-full flex-row items-center px-4 shadow-md"
      style={{
        elevation: 5, // Android shadow
        shadowColor: "#000", // iOS shadow
        shadowOffset: { width: 0, height: 2 },
        shadowOpacity: 0.15,
        shadowRadius: 4,
      }}
    >
      <MaterialIcons name="search" size={24} color="#808080" />
      <TextInput
        placeholder="Where To?"
        placeholderTextColor="#808080"
        className="ml-4 text-[17px] flex-1 text-gray-800"
        editable={false}
      />
    </TouchableOpacity>
  );
};

export default SearchBar;
