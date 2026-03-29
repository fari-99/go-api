# 🌊 Business Flows

This directory contains the core orchestration logic for complex business processes, specifically focusing on state transitions and multi-step transaction flows.

## 📁 Business Flow Components

### 🔄 FSM Transactions (`sm_transactions.go`)
- **Finite State Machine Implementation**: Using `looplab/fsm` to define and manage valid state transitions.
- **Workflow Management**: Automation of business logic as entities move through various statuses (e.g., Pending -> In Progress -> Completed/Failed).
- **Audit Trails**: Built-in logging for every state transition to ensure full traceability.

### 🏠 Property Flow (`properties/`)
- Specialized logic for handling property-related lifecycles, and their corresponding status updates.

### 🏗️ Base Flow (`base.go`)
- Abstractions and shared interfaces to ensure all business flows adhere to a consistent structure.
- Methods for triggering events and handling global state hooks.

---

## 💡 Why FSM?
Using an FSM prevents invalid states and ensures that business logic is explicitly defined and enforced by the code, rather than buried in complex `if-else` blocks.
