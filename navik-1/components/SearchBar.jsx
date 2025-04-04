// components/SearchBar.tsx
import { View, TextInput, TouchableOpacity } from "react-native";
import React from "react";
import { MaterialIcons } from "@expo/vector-icons";
import { useRouter } from "expo-router";

const SearchBar = () => {
  const router = useRouter();
  
  return (
    <TouchableOpacity 
      onPress={() => router.push("/plan-ride")}
      className="bg-[#e6e6e6] h-[7vh] mx-3 mt-10 rounded-full flex-row items-center px-4"
    >
      {/* Search Icon */}
      <MaterialIcons name="search" size={24} color="#808080" />

      {/* Non-editable Text Input */}
      <TextInput
        placeholder="Where To?"
        placeholderTextColor="#808080"
        className="ml-4 text-lg flex-1 text-gray-800"
        editable={false}
      />
    </TouchableOpacity>
  );
};

export default SearchBar;
