## **Product Requirements Document: TWIGGER**

**Version:** 1.6 **Status:** Final **Owner:** Product Team **Date:** September 9 2025

### **Change Log**

* **v1.6 (current):** Re-integrated Non-Functional Requirements and clarified user stories. Refined structure to serve as a comprehensive overview PRD.  
* **v1.5:** Adapted for a solo developer workflow using Go, Google Cloud Run, and GraphQL.  
* **v1.4:** Added multilingual support.

---

### **1\. Introduction**

#### **1.1. Product Vision**

To create an intelligent, AI-powered companion that empowers amateur and hobbyist gardeners to design, cultivate, and manage the garden of their dreams, regardless of location, scale, or environmental conditions.

#### **1.2. Target Audience**

Amateur gardeners and passionate hobbyists across the world, seeking advanced tools for design, planning, and proactive care.

#### **1.3. Business Model**

The application will operate on a trial-to-subscription model. A 14-day free trial provides full access, after which users must subscribe to maintain functionality.

---

### **2\. User Experience & Design (UI/UX)**

#### **2.1. General Principles**

The design shall be clean, intuitive, and data-rich. The user interface must empower users without overwhelming them.

#### **2.2. Navigation**

The application's primary navigation is a bottom tab bar with five distinct sections: **Home**, **Calendar**, **Camera**, **Add (+)**, and **Profile**.

#### **2.3. Accessibility & Localization**

The application must adhere to **WCAG 2.1 Level AA** standards. Full **multilingual support** is mandatory. All user-facing strings must be retrieved from a central translation service (via the GraphQL API) to ensure consistency and avoid duplicated translation files across services.

---

### **3\. Core Features & User Epics**

This section outlines the primary user-facing capabilities. Detailed specifications for each epic will be maintained in separate documents.

* **Epic 0: Membership System**  
  * **Summary:** A user must authenticate to use the app. Registration starts a trial period tied to a user's workspace, enabling future team/family plans.  
  * **User Story:** "As a new user, I want a simple and secure way to sign up, so I can start using the app's features immediately within a free trial period."  
* **Epic 1: The Digital Twin**  
  * **Summary:** Creation of a precise virtual model of the user's garden using maps, AI, and manual drawing tools.  
  * **User Story:** "As a gardener, I want to create an accurate digital map of my garden beds and plants, so I can visualize my space and plan effectively."  
* **Epic 2: The AI Gardener**  
  * **Summary:** An intelligent assistant that provides conversational advice, "what-if" simulations, and proactive tips.  
  * **User Story:** "As an inexperienced gardener, I want to ask questions in plain language and get expert advice on plant placement and care, so I can make better decisions."  
* **Epic 3: The Taskmaster**  
  * **Summary:** A dynamic scheduling system that generates daily, weekly, and monthly tasks based on the user's specific plants and local weather.  
  * **User Story:** "As a busy person, I want the app to automatically tell me what I need to do in my garden and when, so I don't miss important care tasks."  
* **Epic 4: The Digital Encyclopedia**  
  * **Summary:** A comprehensive database of plants, pests, and diseases with a powerful camera-based identification tool.  
  * **User Story:** "As a curious gardener, I want to quickly identify an unknown plant, pest, or disease by taking a photo, so I can learn more and take appropriate action."  
* **Epic 5: The Garden Journal**  
  * **Summary:** A feature for users to document their gardening journey with notes, photos, and tags.  
  * **User Story:** "As a proud gardener, I want to keep a visual diary of my garden's progress with photos and notes, so I can look back on my successes and share them with friends."  
* **Epic 6: The Algorithmic Home Feed**  
  * **Summary:** The primary screen of the application, providing a dynamic and personalized stream of information.  
  * **User Story:** "As a user, I want a centralized home screen that shows me the most relevant and timely information about my garden at a glance, so I can stay informed and engaged without having to hunt for it."

---

### **4\. System Architecture & Technology**

#### **4.1. High-Level Architecture**

The system is a set of **Go-based services** on **Google Cloud Run**. The Flutter client communicates with a primary **GraphQL gateway** service. Asynchronous events are handled via Pub/Sub.

#### **4.2. Definitive Technology Stack**

* **Mobile App:** Flutter  
* **Backend Language:** **Go (Golang)**  
* **Backend Services:** **Google Cloud Run**  
* **API Layer:** **GraphQL** (default), with simple REST/JSON for webhooks.  
* **Database:** PostgreSQL 15+ with PostGIS 3+ on **Google Cloud SQL**.  
* **Authentication:** **Firebase Authentication** (client-side), **Cloud Run built-in IAM** (service-to-service).  
* **Asynchronous Tasks:** Google Cloud Pub/Sub and Cloud Tasks.

---

### **5\. Non-Functional Requirements**

These are critical quality attributes that the system must satisfy.

#### **5.1. Performance**

* The garden map interface must be highly responsive with fluid interactions.  
* All network requests must be asynchronous to prevent UI blocking.  
* Backend services must be designed to scale efficiently, with a target of handling **10,000 concurrent users**.

#### **5.2. Offline & Cross-Device Synchronization**

* Core data (garden model, tasks, journal entries) **must be cached on-device** using SQLite for offline access.  
* The system must implement a robust synchronization mechanism to ensure data is consistent across all of a user's logged-in devices.

#### **5.3. Security & Data Privacy**

* The application must comply with relevant data protection laws (e.g., GDPR).  
* Explicit user consent must be obtained during onboarding before any personal data is collected or shared with third-party services.  
* All inter-service communication will be secured using Google's built-in IAM authentication.

---

### **6\. Solo Developer Workflow & Guiding Principles**

This project is optimized for a single developer to achieve maximum velocity.

#### **6.1. Development Environment: Cloud-Native First**

Local emulation is **eliminated**. All development will occur against a real Google Cloud project with `dev-` prefixed resources. The primary IDE will be Google Cloud Shell Editor or GitHub Codespaces.

#### **6.2. Claude Code as Pair Programmer**

The AI assistant is the primary tool for code generation, testing, and review. A `CLAUDE_CONTEXT.md` file will be maintained at the repo root to provide consistent context for all interactions.

#### **6.3. Code & Repository Structure**

Simplicity is key.

* **Structure:** Organize code by feature (e.g., `services/users/`), not by technical layer.  
* **Migrations:** All SQL migrations are stored in a single `migrations/` folder at the repo root and applied via `golang-migrate` during deployment.

#### **6.4. Practices to Avoid**

* ❌ **Local Testing:** Test directly against the dev GCP environment.  
* ❌ **Initial Unit Tests:** Focus on integration tests. Add unit tests for stable, complex logic only.  
* ❌ **Complex Branching:** Use trunk-based development on `main` with feature flags.  
* ❌ **Formal API Docs:** The GraphQL schema is the documentation.  
* ❌ **Complex Microservice Patterns:** Use simple, direct HTTP/GraphQL calls.  
* ❌ **Complex DI Frameworks:** Keep it simple.

---

### **7\. Lean Testing Strategy**

The goal of testing is to ensure core features are stable without slowing down development.

* **Primary Focus:** **Integration tests** that exercise the real services deployed in the `dev` GCP environment.  
* **Test Case Generation:** Use Claude to generate test scenarios and boilerplate code from the PRD requirements.

---

### **8\. CI/CD Pipeline**

The pipeline is designed for speed and simplicity.

* **Source Control:** GitHub, with all work committed directly to the `main` branch.  
* **Platform:** **Google Cloud Build**.  
* **Workflow:** Every commit to `main` is automatically tested and deployed to the `dev` environment. Promotion to `production` is a **manual approval step**.

