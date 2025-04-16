import json
import random
import string

def generate_random_users(num_users=1000):
    first_names = ["Ajay", "Vijay", "Rahul", "Sunil", "Anil", "Amit", "Sumit", "Rajesh", "Ramesh", 
                  "Priya", "Neha", "Pooja", "Deepa", "Shreya", "Anjali", "Kavita", "Sunita"]
    
    last_names = ["Sharma", "Verma", "Singh", "Patel", "Gupta", "Kumar", "Jain", "Agarwal", "Mishra", 
                 "Yadav", "Chauhan", "Soni", "Khanna", "Kapoor", "Chopra", "Mehta", "Shah"]
    
    user_types = ["driver", "customer"]
    
    users = []
    
    for i in range(num_users):
        first_name = random.choice(first_names)
        last_name = random.choice(last_names)
        user_type = random.choice(user_types)
        
        email = f"{user_type.lower()}-{first_name.lower()}{i}@example.com"
        
        phone = "+91" + ''.join(random.choices(string.digits, k=10))
        
        user = {
            "email": email,
            "password": "SecurePass123!",
            "first_name": first_name,
            "last_name": last_name,
            "user_type": user_type,
            "phone": phone
        }
        
        users.append(user)
    
    # Save users to a JSON file
    with open("random_users.json", "w") as f:
        json.dump(users, f, indent=2)
    
    print(f"{num_users} random users generated and saved to random_users.json")

if __name__ == "__main__":
    generate_random_users(1000) 