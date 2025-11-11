---
trigger: always_on
---

* * *

👑 Master Prompt VPN Automation System (Iran Focus)
====================================================

You are the expert systems architect and implementation generator. Build a **production-ready, high-performance, decoupled SvelteKit + Go Monorepo** for a secure VPN Automation Admin Tool. The system must adhere to all finalized requirements, including the "Bank Manager" resource model and the "Globe-centric" UI.

1\. ⚙️ Core Architecture & Tech Stack
-------------------------------------

### 1.1 Backend (Go: Control Plane)

*   **Language/Framework:** **Go** (for high-concurrency) using **Chi** or **Fiber** as the HTTP router.
    
*   **Database:** **PostgreSQL** with **GORM** (Prisma is forbidden).
    
*   **Routing:** The backend serves the static SvelteKit files and exposes all JSON APIs under `/api/v1/`.
    
*   **Authentication:** Implement a **self-hosted JWT** system in Go. Firebase/Auth0 are forbidden. Enforce the **"Single Concurrent Session"** rule.
    
*   **Agent Communication:** Must establish secure communication channels:
    
    *   **Linux Agent:** Secure **WebSocket/HTTP(S)** with a **Per-Agent Token**.
        
    *   **RouterOS:** **SSH/RouterOS API** (Port 8729 preferred) for Provision/Cleanup scripts.
        

### 1.2 Frontend (SvelteKit: SPA)

*   **Architecture:** **SvelteKit** built as a **Static Site (SSG)/Single Page Application (SPA)**. Zero server-side rendering or logic.
    
*   **Components:** Must use **shadcn-svelte** for core components (Card, Button, Table, Tabs, Modal, Toast).
    
*   **Styling:** **Tailwind CSS**.
    

2\. 🛡️ Resource Management (Smart IPAM/PortAM)
-----------------------------------------------

Implement the **"Bank Manager" (3-State Leasing)** model with the following rules:

*   **Status Model:** Use the three states: `Available`, `Reserved`, `Used`.
    
*   **IPAM/PortAM Governance:**
    
    *   Implement **database row-level locks** (`FOR UPDATE`) during allocation to prevent race conditions.
        
    *   **CIDR/Range Validation:** Prevent the creation of new pools if their CIDR or range overlaps with any existing pool.
        
    *   **Unsafe Ports:** Implement a **Port Blacklist** to prevent allocation of sensitive ports (e.g., 22, 80, 443, 8728).
        
    *   **State-Aware Allocation:** Before reserving a Port, the backend **must** check the latest Agent Heartbeat JSONB (`used_ports`) to prevent conflicts with local services.
        
*   **Recycle Model (The "Bailiff"):** Implement **Active, Command-Triggered Recycle**. When an admin runs `Provision`, `Re-provision`, or `Cleanup`, the Go backend (Bailiff) _must_ atomically reclaim (Recycle) any `Reserved` or `Used` resources associated with that device in the database _before_ proceeding with the new job.
    
*   **Lease Workflow (Agent-Claim):** The agent _never_ writes to the database. Upon job success, the agent **must** call `POST /api/v1/ipam/claim` with a `lease_token` and `proofs` (e.g., `wg show` output). The backend verifies the claim and moves the status to `Used`.
    

3\. 🎨 UI/UX & Visual Identity (SvelteKit Implementation)
---------------------------------------------------------

### 3.1 Global Style & Layout

*   **Theme:** **Dark Mode** with **"Cool Mode"** aesthetic.
    
*   **Background:** Global `Linear Gradient Dots` background.
    
*   **Readability:** All informational content _must_ be contained within **Glassmorphism `Card`s** (using `backdrop-blur` on a semi-transparent background) to ensure `text-bright-1/2/3` readability against the animated background.
    
*   **Navigation (Unified Dock):**
    
    *   **Header and `?tab=` are forbidden**.
        
    *   The primary navigation _must_ be a single, unified **Magic UI `Dock`** component fixed to the top-center.
        
    *   The `Dock` must include: `LogoBadge`, Nav Links (`Dashboard`, `Devices`, `Tunnels`, `Ops`), and the `Settings` modal trigger icon.
        

### 3.2 Routing (File-Based Structure)

Implement the following file-based routes in SvelteKit:

*   `/` (Dashboard)
    
*   `/devices` (Device List)
    
*   `/devices/[id]` (Device Details)
    
*   `/tunnels` (Tunnel List & Builder)
    
*   `/ops` (Jobs & Logs Timeline)
    
*   `/auth/login` (Login Page)
    
*   `/+error` (404 Page)
    

### 3.3 Core Page Component Requirements

| Route | Primary Visual & Function | Data/Action |
| --- | --- | --- |
| /auth/login | Lightweight:OnlyLottie(tunnel animation) and the login form. TheGlobeis forbidden here. | On success, redirect to/and show aToast Notification. |
| /(Dashboard) | Globe-Centric:Globe (cobe.js)is the full-page canvas. | Rendersblinking pinpointsfor devices. KPIs must be Glassmorphism Cards floatingon topof the globe. |
| /devices | Contextual Animation:Uses theCardHoverEffectfor the list of devices. | Clicking navigates to/devices/[id]. |
| /devices/[id] | Device Detail Page. | Must contain action buttons:Provision,Re-provision(smart fix), andCleanup(total wipe). |
| /tunnels | Tunnel List andMulti-step Wizardfor creating tunnels. | Protocols: WireGuard, Hysteria2, V2Ray/Reality, L2TP, OpenVPN, SSR, Vless. TheConfirmstep triggers the "Bank Loan" workflow. |
| /ops | Direct Timeline:Must render theFeatureProgresstimeline directly on the page (not in a modal). | Fetches and displays job history and logs. |
| Settings(Modal) | Modal triggered by Dock icon. UsesTabsforCloudflare (ACME),IPAM, andPortAMconfiguration. | Inputs must useInputGradientBorder. |

Export to Sheets

4\. 🌐 External Integration & GeoIP
-----------------------------------

*   **Primary GeoIP:** Use **`ipapi.ipspeed.info`** (for quick flag and geo data).
    
*   **Fallback GeoIP:** If Primary fails, use **`findip.net`**.
    
*   **Ping Nodes:** The backend must use **`ir1.node.check-host.net`** (Primary) and **`ir6.node.check-host.net`** (Fallback) for latency probes.

---