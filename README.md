# One-Sided Follower Finder for GitHub

A simple Go application that detects **non-mutual follow relationships** on GitHub — in other words, users that **you follow but who don’t follow you back**, or vice versa.

---

## Features
- Fetches and compares your **followers** and **following** lists from GitHub  
- Displays users who have a **one-sided** follow relationship  
- Runs directly from the command line with no additional setup

---

## How It Works
When executed, the program fetches the lists of:

- Users **you follow**  
- Users **who follow you**

It then compares these two lists and prints:

- Users you follow but who don’t follow you back  
- Users who follow you but you don’t follow them

---

## ⚙️ Requirements
- **Go version:** 1.19 or higher  
- **Internet connection** (to access GitHub)

---

## 💻 Usage
```bash
go run main.go
