# Agent Instruction Configuration

## Purpose
This document outlines the behavior, rules, and guidelines for AI agents interacting with the `taller-gestion` project. It ensures consistency, reliability, and adherence to project requirements.

---

## General Guidelines
1. **Accuracy**: Agents must provide accurate and context-aware responses based on the project's codebase and documentation.
2. **Relevance**: Responses should be directly related to the user's query and the project's scope.
3. **Clarity**: All outputs must be clear, concise, and free of unnecessary jargon.
4. **Security**: Avoid exposing sensitive information, such as API keys, credentials, or private data.

---

## Behavior Rules
1. **Language**: Use English for all responses unless specified otherwise.
2. **Tone**: Maintain a professional and neutral tone.
3. **Error Handling**: Provide clear explanations for errors and suggest actionable solutions.
4. **Code Generation**:
   - Ensure code is functional and adheres to best practices.
   - Include comments where necessary for clarity.
5. **Prohibited Actions**:
   - Do not generate or suggest malicious code.
   - Avoid speculative or unsupported claims.

---

## Interaction Guidelines
1. **User Queries**:
   - Respond to technical questions related to JavaScript, React, SQL, and npm.
   - Request clarification if the query is ambiguous.
2. **Code Context**:
   - Use the provided code context to tailor responses.
   - Avoid assumptions beyond the given context.
3. **Documentation**:
   - Reference project-specific documentation when applicable.
   - Suggest updates to documentation if gaps are identified.
4. **Beginer-Friendly**: Provide explanations suitable for users with varying levels of expertise.
---

## Testing and Validation
1. **Code Testing**:
   - Ensure generated code is testable and aligns with the project's testing strategy.
   - Recommend unit tests for new components.
2. **Validation**:
   - Verify that responses align with the project's architecture and dependencies.
3. **Code Coverage:
   - Aim for a minimum of 70% code coverage in generated code.
   - Suggest SQL mocks for repository testing when applicable.
---

## Code Generation Guidelines
1. **Best Practices**: Follow industry best practices for code structure, naming conventions, and documentation.
2. **Modularity**: Encourage modular code design for maintainability and scalability.
3. **Error Handling**: Implement robust error handling and logging mechanisms.
4. **Performance**: Optimize code for performance and efficiency.
5. **Descriptive**: Use descriptive variable and function names to enhance readability.
6. **Comments**: Include comments to explain complex logic or decisions in the code.
7. **Testing**: Recommend unit tests and integration tests for new code components.
---

## Git
1. **Commit Messages**: Use clear and descriptive commit messages that summarize changes.
2. **Branching Strategy**: Follow a consistent branching strategy (e.g., feature branches, development, main).
3. **Changes as new branches**: When making changes, create a new branch for each feature or bug fix. Avoid committing directly to the main branch.

## Notes
- This configuration is subject to updates as the project evolves.
- Feedback from developers will be used to refine agent behavior.