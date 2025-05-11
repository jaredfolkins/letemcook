# Let'em Cook (LEMC) – Overview and Design

## Executive Summary

LEMC (Let'em Cook) is an open-source tool enabling developers to automate and execute scripts ("recipes") with real-time web UI feedback. Its core philosophy is **developer-centric operations** ("Ops your Devs"), empowering teams to manage their own operational tasks directly. Key principles include:

*   **Simplicity & Script-First:** Prioritizes plain scripts (Bash, Python, etc.) over complex DSLs or UIs, aligning with rapid development and AI-assisted coding.
*   **Containerization for Consistency:** Leverages Docker to ensure recipes run reliably in isolated, reproducible environments.
*   **Accessibility:** Makes operational tasks accessible via a simple web interface, removing silos and enabling quicker execution of common procedures.
*   **Empowerment:** Follows the "You build it, you run it" mantra, encouraging developers to own the full lifecycle of their services, including operational aspects.

This document explores the design choices, influences, and the cultural context LEMC aims to serve.

**Let'em Cook (LEMC)** is an open-source tool that lets developers automate and execute predefined "recipes" (scripts) on demand, with results streamed live to a web interface. It aims to "Ops your Devs," meaning it empowers developers to perform operational tasks themselves, rather than relying on a separate DevOps team. LEMC addresses the scenario where important scripts or glue code are siloed with individual engineers or running ad-hoc on someone's machine. By packaging such scripts into containers and providing a UI to run them, LEMC makes these tasks accessible to the whole team. The core idea is to allow even non-specialist team members – potentially even a manager or customer – to click a button and "just do the thing" that a script would do, in a safe, repeatable way.

## Philosophical Underpinnings of Core Features

LEMC's architecture is a direct reflection of its philosophy. Instead of detailing each feature here (see [Key Components](FEATURES.md#key-components) and [Execution Environment](FEATURES.md#execution-environment) for that), this section explains the *why* behind key design choices:

*   **Cookbooks and Recipes:** The organization into cookbooks and recipes (see [Key Components](FEATURES.md#key-components) for details) is designed for clarity and reusability, allowing teams to build a library of automated tasks. The lightweight permission model supports this by enabling controlled sharing.
*   **Script-Based Steps:** The choice of plain scripts as the foundation for recipe steps (detailed in [Execution Environment](FEATURES.md#execution-environment)) is central to LEMC. This "language-first" approach maximizes developer familiarity, flexibility, and allows easy integration of existing automation scripts with minimal changes. It avoids vendor lock-in to a specific DSL and embraces the power of general-purpose programming languages.
*   **Containerized Execution:** Using Docker for execution (see [Execution Environment](FEATURES.md#execution-environment) for details) provides critical benefits: **consistency** (eliminating "works on my machine" issues), **isolation** (preventing interference between tasks or with the host), and **dependency management** (packaging all necessary tools within the image). This aligns with modern DevOps best practices.
*   **Real-Time Feedback via Web UI:** The verb-based system for real-time UI updates (see [Real-Time UI Communication (LEMC Verbs)](FEATURES.md#real-time-ui-communication-lemc-verbs) for details) is designed to provide immediate visibility into script execution. This transparency is crucial for debugging, monitoring, and building user confidence, transforming scripts from black boxes into interactive processes.
*   **Scheduling:** The inclusion of scheduling capabilities (see [Scheduling](FEATURES.md#scheduling) for details) extends LEMC from on-demand execution to proactive, automated maintenance and operational tasks, replacing potentially fragile cron jobs with managed, observable processes.
*   **Tech Stack Choices:** The Go backend, SQLite database, and HTMX frontend (see [Tech Stack Highlights](FEATURES.md#tech-stack-highlights) for details) were chosen for **simplicity, performance, and ease of deployment**. The goal was a self-contained system that is easy to run and maintain, even for small teams.

## Security Philosophy

LEMC's security model is intentionally **lightweight and pragmatic**, designed for trusted environments like internal development teams. The core tenets are:

*   **Container Isolation as Primary Defense:** The primary security mechanism is Docker container isolation (see [Security](FEATURES.md#security) for details). This sandboxing is leveraged to confine script execution and limit potential impact on the host system or other operations. This approach assumes that access to LEMC and its underlying Docker socket is already controlled.
*   **Simplified User Management:** The user model (admin creation on first launch, basic user management) is designed for ease of use in small to medium-sized teams where complex enterprise-grade RBAC would be overkill. (Details in [Security](FEATURES.md#security)).
*   **Trust in Developer-Defined Recipes:** LEMC operates on the principle that recipes are created or vetted by the team. The focus is on providing a safe *execution environment* for these trusted scripts, rather than on policing the scripts themselves.

This philosophy prioritizes developer agility and self-service within a trusted context, relying on established container security features.

## "Own Your Ops" Philosophy

LEMC's architecture (simple scripts, containers, minimal config) is meant to flatten this – ideally, the same developers writing application code can also write and run the ops recipes, keeping the loop tight. It's inspired by the idea that in modern teams (especially with AI assistance accelerating coding), a small team can ship faster if they handle their own operational needs. This directly echoes the DevOps mantra "**You build it, you run it**," famously advocated by Amazon's CTO Werner Vogels in 2006. By integrating operations into the development workflow (in this case, via a handy internal tool), LEMC tries to reduce friction and eliminate the scenario of throwing code "over the wall" to ops teams.

## Key Influences and Design Precedents

LEMC builds upon established concepts in system design, automation, and software practices. Understanding these influences provides context for its design choices.

### 1. Job Scheduling & Automation (Cron, CI/CD, Runbook Automation)

*   **Core Idea:** Automating predefined tasks is a foundational concept, from Unix `cron` (1975) for scheduled jobs to modern CI/CD systems (like Jenkins) for event-driven pipelines.
*   **LEMC's Angle:** Provides a user-friendly UI for on-demand and scheduled execution of containerized scripts, akin to how Jenkins offers a "Build Now" button or Rundeck provides self-service runbook automation. It simplifies the "just run my script" need with modern tooling.

### 2. Isolation & Containerization (Chroot, Docker)

*   **Core Idea:** Running processes in isolated, reproducible environments has evolved from `chroot` (1979) and FreeBSD Jails to modern OS-level virtualization like Docker (2013).
*   **LEMC's Angle:** Fully embraces Docker to package recipes with their dependencies. This ensures scripts run consistently and safely, isolated from the host and each other, leveraging decades of OS-level isolation advancements.

### 3. Developer-Centric Operations (DevOps, IaC)

*   **Core Idea:** The DevOps movement, particularly the "You build it, you run it" principle, encourages developers to take operational responsibility. Infrastructure as Code (IaC) tools (like Chef, with its "cookbooks" and "recipes") allow ops tasks to be codified.
*   **LEMC's Angle:** Directly embodies this by enabling developers to write, manage, and execute their own ops scripts. It can be seen as "Operations as Micro-Code," using familiar terminology but for smaller, self-contained tasks.

### 4. Real-time Feedback & Interaction (Actor Model, Live Logging)

*   **Core Idea:** Systems providing real-time insight (e.g., live log streaming via `tail -f`, Jupyter Notebooks for interactive output) enhance usability. The Actor Model (1973) provides a conceptual basis for isolated components communicating via messages.
*   **LEMC's Angle:** Uses a verb-based system over WebSockets to stream HTML, CSS, and JS from running scripts to the UI. This provides live, rich feedback, making script execution transparent and interactive, conceptually similar to an actor sending status messages.

These historical and conceptual pillars are synthesized in LEMC to create a pragmatic tool for modern development teams.

## References

* LEMC Documentation and README (Jared Folkins, 2025) – for descriptions of LEMC features and philosophy.
* Unix Cron history – job scheduling in 1975.
* IBM JCL Basics – multi-step jobs in mainframe systems.
* Werner Vogels (Amazon) interview, 2006 – "You build it, you run it" DevOps principle.
* Progress Chef (formerly Chef) documentation – use of "recipes" and "cookbooks" in 2000s config management.
* Aqua Security: *History of Containers* – evolution from chroot (1979) to Docker (2013).
* GeeksforGeeks: Sandbox Security Model – on sandboxing untrusted code (Java applet model).
* Rundeck project intro – self-service runbook automation with web UI (2010s).
* Penn State Univ. – Jupyter Notebook overview – mixing code and output in one interface (2010s).
* Atlassian Blog on DevOps/ChatOps – context on DevOps collaboration and ChatOps trend.
