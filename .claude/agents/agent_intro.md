## When to Use Specialized Agents in Claude Code

### 🎯 **Specialized Agents Are Valuable When:**

#### 1. **Complex Domain-Specific Problems**
```markdown
# PostGIS Spatial Analysis Agent
You are a PostGIS expert specializing in garden spatial analysis.
Focus: Optimizing spatial queries, coordinate systems, geometry validation
Knowledge: All ST_* functions, GIST indexing, projection transforms
```
**Use when:** Writing complex spatial queries, optimizing PostGIS performance

#### 2. **Algorithm Implementation**
```markdown
# Shade Calculation Agent
You are an expert in solar position algorithms and shadow projection.
Focus: Sun path calculation, shadow geometry, seasonal variations
Knowledge: Solar equations, 3D geometry, temporal calculations
```
**Use when:** Implementing Part 4 (Analysis Engine)

#### 3. **Performance Optimization**
```markdown
# Database Performance Agent
You are a PostgreSQL performance tuning expert.
Focus: Query optimization, index strategies, EXPLAIN ANALYZE interpretation
Tools: pg_stat_statements, query planning, connection pooling
```
**Use when:** Queries are slow, need indexing strategy

#### 4. **Security Reviews**
```markdown
# Security Audit Agent
You are a security expert focusing on API and database security.
Focus: SQL injection, auth bypass, data leakage, OWASP top 10
Mindset: Think like an attacker, find vulnerabilities
```
**Use when:** Reviewing auth implementation, API endpoints

### 📊 **Decision Matrix: Regular vs Specialized Agent**

| Scenario                        | Regular Claude | Specialized Agent | Why                   |
| ------------------------------- | -------------- | ----------------- | --------------------- |
| CRUD operations                 | ✅              | ❌                 | Standard patterns     |
| PostGIS spatial queries         | ❌              | ✅                 | Deep expertise needed |
| Basic API setup                 | ✅              | ❌                 | Common knowledge      |
| Shade calculation algorithm     | ❌              | ✅                 | Complex domain        |
| Docker setup                    | ✅              | ❌                 | Well-documented       |
| GraphQL DataLoader optimization | ❌              | ✅                 | Specific expertise    |
| Writing tests                   | ✅              | ❌                 | Standard practice     |
| Firebase auth integration       | ✅              | ❌                 | Good documentation    |
| Database performance tuning     | ❌              | ✅                 | Specialized knowledge |

### 🚀 **Specialized Agents for Your Project**

#### **Agent 1: PostGIS Spatial Expert**
```markdown
# When to activate: Parts 3 & 4
You are a PostGIS expert for garden mapping.

Expertise:
- Spatial relationship queries (ST_Contains, ST_Intersects)
- Coordinate transformations (SRID 4326 ↔ local projections)
- Geometry validation and repair
- Spatial index optimization

Always:
- Validate geometries before storage
- Use geography type for measurements
- Consider GIST indexes
- Explain spatial function choices
```

#### **Agent 2: GraphQL Optimization Specialist**
```markdown
# When to activate: Part 6
You are a GraphQL performance expert using gqlgen.

Focus:
- DataLoader implementation for N+1 prevention
- Query complexity calculation
- Resolver optimization
- Subscription efficiency

Always:
- Batch database calls
- Implement field-level caching
- Limit query depth
- Use context for cancellation
```

#### **Agent 3: Plant Database Domain Expert**
```markdown
# When to activate: Part 2
You are a botanist and data architect for plant databases.

Knowledge:
- Plant taxonomy and relationships
- Companion planting rules
- Growing condition requirements
- Multi-source data reconciliation

Consider:
- Conflicting data between sources
- Seasonal variations
- Regional differences
- Cultivar vs species distinctions
```

### 🎨 **Creating Effective Specialized Agents**

#### **Template Structure:**
```markdown
# [Agent Name]
You are a [specific expertise] specialist.

## Core Competencies
- [Specific skill 1]
- [Specific skill 2]

## Constraints
- [Limitation 1]
- [Limitation 2]

## Always
- [Best practice 1]
- [Best practice 2]

## Never
- [Anti-pattern 1]
- [Anti-pattern 2]

## Output Format
[Specific format requirements]

## Context
Working on: [current task]
Project: [project description]
Stack: [tech stack]
```

### 📈 **Signs You Need a Specialized Agent**

1. **Repeated Mistakes** in specific domain
2. **Inconsistent Approaches** to similar problems
3. **Lack of Domain Knowledge** showing in outputs
4. **Complex Algorithm Needs** beyond general knowledge
5. **Performance Issues** requiring deep optimization

### ⚡ **Quick Specialized Agents** (For Your Project)

```bash
# For complex PostGIS work
"Act as a PostGIS expert. Focus only on spatial SQL optimization."

# For test writing
"You are a Go testing expert. Write comprehensive table-driven tests."

# For API design
"You are a REST/GraphQL API architect. Focus on clean, consistent design."

# For error handling
"You are a Go error handling expert. Implement proper error wrapping and context."
```

### 🔄 **Agent Switching Strategy**

```markdown
1. Start with regular Claude for exploration
2. Identify complexity/expertise need
3. Switch to specialized agent for that component
4. Document solution in code/comments
5. Return to regular Claude for integration
```

### 💡 **For Solo Developer Workflow**

Since you're working alone with Claude Code:

**Create an `agents/` directory:**
```
project/
├── agents/
│   ├── postgis-expert.md
│   ├── graphql-specialist.md
│   ├── performance-tuner.md
│   └── security-auditor.md
```

**Usage pattern:**
```markdown
/clear
[paste content from agents/postgis-expert.md]
"Optimize this garden boundary query: ..."
```

### 🎯 **ROI: When Agents Save Most Time**

**HIGH VALUE** (Definitely use agent):
- PostGIS spatial operations (Part 3, 4)
- Shade/drainage algorithms (Part 4)
- GraphQL DataLoader setup (Part 6)
- Database performance tuning

**MEDIUM VALUE** (Consider agent):
- Authentication setup (Part 7)
- Caching strategies
- API design decisions

**LOW VALUE** (Regular Claude fine):
- CRUD operations
- Basic testing
- Documentation
- File organization

The key insight: **Use specialized agents when the problem requires deep expertise that would take you significant research time to acquire**. For your project, this is especially true for PostGIS spatial operations and garden analysis algorithms.