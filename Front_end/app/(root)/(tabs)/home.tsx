import { Text } from "react-native"; // Importing components from React Native
import { SafeAreaView } from "react-native-safe-area-context";
const Home = () => {
  // Functional component
  return (
    <SafeAreaView>
      {" "}
      {/* A container for layout */}
      <Text>Home Page</Text> {/* Displaying text */}
    </SafeAreaView>
  );
};

export default Home; // Exporting the component
