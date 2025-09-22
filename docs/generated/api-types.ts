// Generated TypeScript interfaces from OpenAPI specification
// Version: 1.0.0
// Generated on: auto-generated

/**
 * Product Requirements Management API
 * Comprehensive API for managing product requirements through hierarchical structure of Epics → User Stories → Requirements. 
Features include full-text search, comment system, relationship mapping, and configurable workflows.

 */

// Base API Configuration
export interface ApiConfig {
  baseUrl: string;
  apiKey?: string;
  timeout?: number;
}

// Standard API Response wrapper
export interface ApiResponse<T = any> {
  data?: T;
  error?: {
    code: string;
    message: string;
  };
}

// Pagination wrapper for list responses
export interface ListResponse<T> {
  data: T[];
  total_count: number;
  limit: number;
  offset: number;
}

// Error response format
export interface ErrorResponse {
  error: {
    code: string;
    message: string;
  };
}

// Authentication types
export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  expires_at: string;
  user: User;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

// User types
export interface User {
  id: string;
  username: string;
  email: string;
  role: 'Administrator' | 'User' | 'Commenter';
  created_at: string;
  updated_at: string;
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  role: 'Administrator' | 'User' | 'Commenter';
}

export interface UpdateUserRequest {
  username?: string;
  email?: string;
  role?: 'Administrator' | 'User' | 'Commenter';
}

// Epic types
export type EpicStatus = 'Backlog' | 'Draft' | 'In Progress' | 'Done' | 'Cancelled';
export type Priority = 1 | 2 | 3 | 4;

export interface Epic {
  id: string;
  reference_id: string;
  title: string;
  description?: string;
  status: EpicStatus;
  priority: Priority;
  creator_id: string;
  assignee_id?: string;
  created_at: string;
  last_modified: string;
  
  // Optional populated fields
  creator?: User;
  assignee?: User;
  user_stories?: UserStory[];
  comments?: Comment[];
}

export interface CreateEpicRequest {
  title: string;
  description?: string;
  priority: Priority;
  creator_id: string;
  assignee_id?: string;
}

export interface UpdateEpicRequest {
  title?: string;
  description?: string;
  priority?: Priority;
  assignee_id?: string;
}

// User Story types
export type UserStoryStatus = 'Backlog' | 'Draft' | 'In Progress' | 'Done' | 'Cancelled';

export interface UserStory {
  id: string;
  reference_id: string;
  title: string;
  description?: string;
  status: UserStoryStatus;
  priority: Priority;
  epic_id: string;
  creator_id: string;
  assignee_id?: string;
  created_at: string;
  last_modified: string;
  
  // Optional populated fields
  epic?: Epic;
  creator?: User;
  assignee?: User;
  acceptance_criteria?: AcceptanceCriteria[];
  requirements?: Requirement[];
  comments?: Comment[];
}

export interface CreateUserStoryRequest {
  title: string;
  description?: string;
  priority: Priority;
  epic_id: string;
  assignee_id?: string;
}

export interface UpdateUserStoryRequest {
  title?: string;
  description?: string;
  priority?: Priority;
  assignee_id?: string;
}

// Acceptance Criteria types
export interface AcceptanceCriteria {
  id: string;
  reference_id: string;
  description: string;
  user_story_id: string;
  author_id: string;
  created_at: string;
  last_modified: string;
  
  // Optional populated fields
  user_story?: UserStory;
  author?: User;
  requirements?: Requirement[];
  comments?: Comment[];
}

export interface CreateAcceptanceCriteriaRequest {
  description: string;
  user_story_id: string;
}

export interface UpdateAcceptanceCriteriaRequest {
  description?: string;
}

// Requirement types
export type RequirementStatus = 'Draft' | 'Active' | 'Obsolete';

export interface Requirement {
  id: string;
  reference_id: string;
  title: string;
  description?: string;
  status: RequirementStatus;
  priority: Priority;
  user_story_id: string;
  acceptance_criteria_id?: string;
  type_id: string;
  creator_id: string;
  assignee_id?: string;
  created_at: string;
  last_modified: string;
  
  // Optional populated fields
  user_story?: UserStory;
  acceptance_criteria?: AcceptanceCriteria;
  type?: RequirementType;
  creator?: User;
  assignee?: User;
  source_relationships?: RequirementRelationship[];
  target_relationships?: RequirementRelationship[];
  comments?: Comment[];
}

export interface CreateRequirementRequest {
  title: string;
  description?: string;
  priority: Priority;
  user_story_id: string;
  acceptance_criteria_id?: string;
  type_id: string;
  assignee_id?: string;
}

export interface UpdateRequirementRequest {
  title?: string;
  description?: string;
  priority?: Priority;
  assignee_id?: string;
}

// Comment types
export type EntityType = 'epic' | 'user_story' | 'acceptance_criteria' | 'requirement';

export interface Comment {
  id: string;
  content: string;
  entity_type: EntityType;
  entity_id: string;
  author_id: string;
  parent_comment_id?: string;
  is_resolved: boolean;
  linked_text?: string;
  text_position_start?: number;
  text_position_end?: number;
  created_at: string;
  updated_at: string;
  
  // Optional populated fields
  author?: User;
  parent_comment?: Comment;
  replies?: Comment[];
}

export interface CreateCommentRequest {
  content: string;
  parent_comment_id?: string;
}

export interface CreateInlineCommentRequest {
  content: string;
  linked_text: string;
  text_position_start: number;
  text_position_end: number;
}

export interface UpdateCommentRequest {
  content: string;
}

// Inline comment validation types
export interface InlineCommentValidationRequest {
  comments: InlineCommentPosition[];
}

export interface InlineCommentPosition {
  comment_id: string;
  text_position_start: number;
  text_position_end: number;
}

export interface ValidationResponse {
  valid: boolean;
  errors: string[];
}

// Configuration types
export interface RequirementType {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface RelationshipType {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface RequirementRelationship {
  id: string;
  source_requirement_id: string;
  target_requirement_id: string;
  relationship_type_id: string;
  created_by: string;
  created_at: string;
  
  // Optional populated fields
  source_requirement?: Requirement;
  target_requirement?: Requirement;
  relationship_type?: RelationshipType;
  creator?: User;
}

export interface CreateRelationshipRequest {
  source_requirement_id: string;
  target_requirement_id: string;
  relationship_type_id: string;
}

// Deletion workflow types
export interface DependencyInfo {
  can_delete: boolean;
  dependencies: DependencyItem[];
  warnings: string[];
}

export interface DependencyItem {
  entity_type: string;
  entity_id: string;
  reference_id: string;
  title: string;
  dependency_type: string;
}

export interface DeletionResult {
  success: boolean;
  deleted_entities: DeletedEntity[];
  message: string;
}

export interface DeletedEntity {
  entity_type: string;
  entity_id: string;
  reference_id: string;
}

// Search types
export interface SearchResult {
  entity_type: EntityType;
  entity_id: string;
  reference_id: string;
  title: string;
  description?: string;
  highlight?: string;
  rank: number;
}

export interface SearchResponse {
  results: SearchResult[];
  total_count: number;
  query: string;
  entity_types: string[];
  limit: number;
  offset: number;
}

export interface SearchSuggestionsResponse {
  titles: string[];
  reference_ids: string[];
  statuses: string[];
}

// Hierarchy types
export interface HierarchyNode {
  entity_type: EntityType;
  entity_id: string;
  reference_id: string;
  title: string;
  status: string;
  children?: HierarchyNode[];
}

export interface EntityPath {
  entity_type: EntityType;
  entity_id: string;
  reference_id: string;
  title: string;
}

// Status management types
export interface StatusChangeRequest {
  status: string;
}

export interface AssignmentRequest {
  assignee_id?: string;
}

// Health check types
export interface HealthCheckResponse {
  status: string;
  reason?: string;
}

// List response types
export interface UserListResponse extends ListResponse<User> {}
export interface EpicListResponse extends ListResponse<Epic> {}
export interface UserStoryListResponse extends ListResponse<UserStory> {}
export interface AcceptanceCriteriaListResponse extends ListResponse<AcceptanceCriteria> {}
export interface RequirementListResponse extends ListResponse<Requirement> {}
export interface CommentListResponse extends ListResponse<Comment> {}
export interface RequirementTypeListResponse extends ListResponse<RequirementType> {}
export interface RelationshipTypeListResponse extends ListResponse<RelationshipType> {}

// API Client interface
export interface ApiClient {
  // Authentication
  login(credentials: LoginRequest): Promise<LoginResponse>;
  getProfile(): Promise<User>;
  changePassword(request: ChangePasswordRequest): Promise<void>;
  
  // User management (Admin only)
  createUser(user: CreateUserRequest): Promise<User>;
  getUsers(params?: { limit?: number; offset?: number }): Promise<UserListResponse>;
  getUser(id: string): Promise<User>;
  updateUser(id: string, user: UpdateUserRequest): Promise<User>;
  deleteUser(id: string): Promise<void>;
  
  // Epics
  createEpic(epic: CreateEpicRequest): Promise<Epic>;
  getEpics(params?: {
    creator_id?: string;
    assignee_id?: string;
    status?: EpicStatus;
    priority?: Priority;
    limit?: number;
    offset?: number;
    include?: string;
  }): Promise<EpicListResponse>;
  getEpic(id: string): Promise<Epic>;
  updateEpic(id: string, epic: UpdateEpicRequest): Promise<Epic>;
  deleteEpic(id: string): Promise<void>;
  changeEpicStatus(id: string, status: StatusChangeRequest): Promise<Epic>;
  assignEpic(id: string, assignment: AssignmentRequest): Promise<Epic>;
  validateEpicDeletion(id: string): Promise<DependencyInfo>;
  deleteEpicComprehensive(id: string): Promise<DeletionResult>;
  
  // User Stories
  createUserStory(userStory: CreateUserStoryRequest): Promise<UserStory>;
  getUserStories(params?: {
    epic_id?: string;
    creator_id?: string;
    assignee_id?: string;
    status?: UserStoryStatus;
    priority?: Priority;
    limit?: number;
    offset?: number;
    include?: string;
  }): Promise<UserStoryListResponse>;
  getUserStory(id: string): Promise<UserStory>;
  updateUserStory(id: string, userStory: UpdateUserStoryRequest): Promise<UserStory>;
  deleteUserStory(id: string): Promise<void>;
  
  // Requirements
  createRequirement(requirement: CreateRequirementRequest): Promise<Requirement>;
  getRequirements(params?: {
    user_story_id?: string;
    acceptance_criteria_id?: string;
    type_id?: string;
    creator_id?: string;
    assignee_id?: string;
    status?: RequirementStatus;
    priority?: Priority;
    limit?: number;
    offset?: number;
    include?: string;
  }): Promise<RequirementListResponse>;
  getRequirement(id: string): Promise<Requirement>;
  updateRequirement(id: string, requirement: UpdateRequirementRequest): Promise<Requirement>;
  deleteRequirement(id: string): Promise<void>;
  
  // Comments
  getComments(entityType: EntityType, entityId: string, params?: {
    limit?: number;
    offset?: number;
  }): Promise<CommentListResponse>;
  createComment(entityType: EntityType, entityId: string, comment: CreateCommentRequest): Promise<Comment>;
  createInlineComment(entityType: EntityType, entityId: string, comment: CreateInlineCommentRequest): Promise<Comment>;
  updateComment(id: string, comment: UpdateCommentRequest): Promise<Comment>;
  deleteComment(id: string): Promise<void>;
  resolveComment(id: string): Promise<Comment>;
  unresolveComment(id: string): Promise<Comment>;
  
  // Search
  search(params: {
    q: string;
    entity_types?: string;
    limit?: number;
    offset?: number;
  }): Promise<SearchResponse>;
  getSearchSuggestions(params: {
    query: string;
    limit?: number;
  }): Promise<SearchSuggestionsResponse>;
  
  // Health checks
  readinessCheck(): Promise<HealthCheckResponse>;
  livenessCheck(): Promise<HealthCheckResponse>;
}

// HTTP client configuration
export interface HttpClientConfig {
  baseURL: string;
  timeout?: number;
  headers?: Record<string, string>;
  interceptors?: {
    request?: (config: any) => any;
    response?: (response: any) => any;
    error?: (error: any) => any;
  };
}

// Utility types for API parameters
export type QueryParams = Record<string, string | number | boolean | undefined>;
export type PathParams = Record<string, string>;
export type RequestHeaders = Record<string, string>;

// API endpoint definitions
export const API_ENDPOINTS = {
  // Authentication
  LOGIN: '/auth/login',
  PROFILE: '/auth/profile',
  CHANGE_PASSWORD: '/auth/change-password',
  
  // User management
  USERS: '/auth/users',
  USER: '/auth/users/{id}',
  
  // Epics
  EPICS: '/api/v1/epics',
  EPIC: '/api/v1/epics/{id}',
  EPIC_USER_STORIES: '/api/v1/epics/{id}/user-stories',
  EPIC_STATUS: '/api/v1/epics/{id}/status',
  EPIC_ASSIGN: '/api/v1/epics/{id}/assign',
  EPIC_VALIDATE_DELETION: '/api/v1/epics/{id}/validate-deletion',
  EPIC_DELETE: '/api/v1/epics/{id}/delete',
  EPIC_COMMENTS: '/api/v1/epics/{id}/comments',
  
  // User Stories
  USER_STORIES: '/api/v1/user-stories',
  USER_STORY: '/api/v1/user-stories/{id}',
  USER_STORY_ACCEPTANCE_CRITERIA: '/api/v1/user-stories/{id}/acceptance-criteria',
  USER_STORY_REQUIREMENTS: '/api/v1/user-stories/{id}/requirements',
  USER_STORY_STATUS: '/api/v1/user-stories/{id}/status',
  USER_STORY_ASSIGN: '/api/v1/user-stories/{id}/assign',
  
  // Requirements
  REQUIREMENTS: '/api/v1/requirements',
  REQUIREMENT: '/api/v1/requirements/{id}',
  REQUIREMENT_RELATIONSHIPS: '/api/v1/requirements/{id}/relationships',
  REQUIREMENT_STATUS: '/api/v1/requirements/{id}/status',
  REQUIREMENT_ASSIGN: '/api/v1/requirements/{id}/assign',
  
  // Search
  SEARCH: '/api/v1/search',
  SEARCH_SUGGESTIONS: '/api/v1/search/suggestions',
  
  // Health
  READY: '/ready',
  LIVE: '/live',
} as const;

export type ApiEndpoint = typeof API_ENDPOINTS[keyof typeof API_ENDPOINTS];
