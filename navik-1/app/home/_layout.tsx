// app/home/_layout.tsx
import { createBottomTabNavigator } from "@react-navigation/bottom-tabs";
import { MaterialIcons } from "@expo/vector-icons";
import HomeScreen from "./index";
import ServicesScreen from "./services/index";
import ActivityScreen from "./activity/index";

const Tab = createBottomTabNavigator();

export default function HomeLayout() {
  return (
    <Tab.Navigator
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: "#000000", // Active tab color (blue)
        tabBarInactiveTintColor: "#A9A9A9", // Inactive tab color (gray)
        tabBarStyle: {
          backgroundColor: "#F8F8F8", // Light gray background
          borderTopWidth: 0, // Remove border at the top of the tab bar
          elevation: 5, // Add shadow for a modern look
          height: 65, // Increase height for better spacing
          paddingBottom: 10, // Add padding for icons and labels
        },
        tabBarLabelStyle: {
          fontSize: 12, // Slightly smaller label text
          fontWeight: "600", // Semi-bold labels
        },
        tabBarIconStyle: {
          marginTop: 5, // Add spacing between icon and label
        },
      }}
    >
      <Tab.Screen
        name="Home"
        component={HomeScreen}
        options={{
          tabBarLabel: "Home",
          tabBarIcon: ({ color }) => (
            <MaterialIcons name="home" size={28} color={color} />
          ),
        }}
      />
      <Tab.Screen
        name="Services"
        component={ServicesScreen}
        options={{
          tabBarLabel: "Services",
          tabBarIcon: ({ color }) => (
            <MaterialIcons name="directions-car" size={28} color={color} />
          ),
        }}
      />
      <Tab.Screen
        name="Activity"
        component={ActivityScreen}
        options={{
          tabBarLabel: "Activity",
          tabBarIcon: ({ color }) => (
            <MaterialIcons name="history" size={28} color={color} />
          ),
        }}
      />
    </Tab.Navigator>
  );
}
