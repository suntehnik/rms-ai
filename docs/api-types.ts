/**
 * TypeScript type definitions for Product Requirements Management API
 * Generated from API documentation - suitable for direct import into client projects
 */

// ============================================================================
// AUTHENTICATION TYPES
// ============================================================================

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  expires_at: string;
  user: UserResponse;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string; // minimum 8 characters
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string; // minimum 8 characters
  role: UserRole;
}

export interface UpdateUserRequest {
  username?: string;
  email?: string;
  role?: UserRole;
}

// ============================================================================
// CORE ENTITY TYPES
// ============================================================================

export type UserRole = 'Administrator' | 'User' | 'Commenter';
export type EntityType = 'epic' | 'user_story' | 'acceptance_criteria' | 'requirement';
export type Priority = 1 | 2 | 3 | 4; // 1=Critical, 2=High, 3=Medium, 4=Low
export type EpicStatus = 'Backlog' | 'Draft' | 'In Progress' | 'Done' | 'Cancelled';
export type UserStoryStatus = 'Backlog' | 'Draft' | 'In Progress' | 'Done' | 'Cancelled';
export type RequirementStatus = 'Draft' | 'Active' | 'Obsolete';

export interface User {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  created_at: string;
  updated_at: string;
}

export interface UserResponse {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  created_at: string;
  updated_at: string;
}

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
  
  // Optional populated fields (when using ?include parameter)
  creator?: User;
  assignee?: User;
  user_stories?: UserStory[];
  comments?: Comment[];
}

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

// ============================================================================
// CONFIGURATION TYPES
// ============================================================================

export interface RequirementType {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
  requirements?: Requirement[];
}

export interface RelationshipType {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
  requirement_relationships?: RequirementRelationship[];
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

export interface StatusModel {
  id: string;
  name: string;
  description?: string;
  entity_type: EntityType;
  is_default: boolean;
  created_at: string;
  updated_at: string;
  
  // Optional populated fields
  statuses?: Status[];
  transitions?: StatusTransition[];
}

export interface Status {
  id: string;
  name: string;
  description?: string;
  color?: string; // Hex color code
  order: number;
  is_initial: boolean;
  is_final: boolean;
  status_model_id: string;
  created_at: string;
  updated_at: string;
  
  // Optional populated fields
  status_model?: StatusModel;
  from_transitions?: StatusTransition[];
  to_transitions?: StatusTransition[];
}

export interface StatusTransition {
  id: string;
  name?: string;
  description?: string;
  from_status_id: string;
  to_status_id: string;
  status_model_id: string;
  created_at: string;
  updated_at: string;
  
  // Optional populated fields
  from_status?: Status;
  to_status?: Status;
  status_model?: StatusModel;
}

// ============================================================================
// REQUEST TYPES
// ============================================================================

export interface CreateEpicRequest {
  title: string;
  description?: string;
  priority: Priority;
  assignee_id?: string;
}

export interface UpdateEpicRequest {
  title?: string;
  description?: string;
  priority?: Priority;
  assignee_id?: string;
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

export interface CreateAcceptanceCriteriaRequest {
  description: string;
  user_story_id: string;
}

export interface UpdateAcceptanceCriteriaRequest {
  description?: string;
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
  content?: string;
}

export interface CreateRelationshipRequest {
  source_requirement_id: string;
  target_requirement_id: string;
  relationship_type_id: string;
}

export interface CreateRequirementTypeRequest {
  name: string;
  description?: string;
}

export interface UpdateRequirementTypeRequest {
  name?: string;
  description?: string;
}

export interface CreateRelationshipTypeRequest {
  name: string;
  description?: string;
}

export interface UpdateRelationshipTypeRequest {
  name?: string;
  description?: string;
}

export interface CreateStatusModelRequest {
  name: string;
  description?: string;
  entity_type: EntityType;
  is_default?: boolean;
}

export interface UpdateStatusModelRequest {
  name?: string;
  description?: string;
  is_default?: boolean;
}

export interface CreateStatusRequest {
  name: string;
  description?: string;
  color?: string;
  order: number;
  is_initial?: boolean;
  is_final?: boolean;
  status_model_id: string;
}

export interface UpdateStatusRequest {
  name?: string;
  description?: string;
  color?: string;
  order?: number;
  is_initial?: boolean;
  is_final?: boolean;
}

export interface CreateStatusTransitionRequest {
  name?: string;
  description?: string;
  from_status_id: string;
  to_status_id: string;
  status_model_id: string;
}

export interface UpdateStatusTransitionRequest {
  name?: string;
  description?: string;
}

export interface StatusChangeRequest {
  status: string;
}

export interface AssignmentRequest {
  assignee_id?: string; // null to unassign
}

export interface DeleteEntityRequest {
  force?: boolean;
}

// ============================================================================
// RESPONSE TYPES
// ============================================================================

export interface ApiResponse<T = any> {
  data?: T;
  error?: {
    code: string;
    message: string;
  };
}

export interface ListResponse<T> {
  data: T[];
  total_count: number;
  limit: number;
  offset: number;
}

export interface EpicListResponse extends ListResponse<Epic> {}
export interface UserStoryListResponse extends ListResponse<UserStory> {}
export interface AcceptanceCriteriaListResponse extends ListResponse<AcceptanceCriteria> {}
export interface RequirementListResponse extends ListResponse<Requirement> {}
export interface CommentListResponse extends ListResponse<Comment> {}
export interface UserListResponse extends ListResponse<UserResponse> {}

export interface RequirementTypeListResponse {
  requirement_types: RequirementType[];
  count: number;
}

export interface RelationshipTypeListResponse {
  relationship_types: RelationshipType[];
  count: number;
}

export interface StatusModelListResponse {
  status_models: StatusModel[];
  count: number;
}

export interface StatusListResponse {
  statuses: Status[];
  count: number;
}

export interface StatusTransitionListResponse {
  transitions: StatusTransition[];
  count: number;
}

// ============================================================================
// SEARCH TYPES
// ============================================================================

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

export interface SearchParams {
  q: string;
  entity_types?: string; // comma-separated
  limit?: number;
  offset?: number;
}

export interface SearchSuggestionsParams {
  query: string;
  limit?: number;
}

// ============================================================================
// HIERARCHY & NAVIGATION TYPES
// ============================================================================

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

export interface HierarchyResponse {
  hierarchy: HierarchyNode[];
}

export interface EntityPathResponse {
  path: EntityPath[];
}

// ============================================================================
// DELETION TYPES
// ============================================================================

export interface DependencyInfo {
  can_delete: boolean;
  dependencies: {
    entity_type: string;
    entity_id: string;
    reference_id: string;
    title: string;
    dependency_type: string;
  }[];
  warnings: string[];
}

export interface DeletionResult {
  success: boolean;
  deleted_entities: {
    entity_type: string;
    entity_id: string;
    reference_id: string;
  }[];
  message: string;
}

// ============================================================================
// QUERY PARAMETERS TYPES
// ============================================================================

export interface EpicListParams {
  creator_id?: string;
  assignee_id?: string;
  status?: EpicStatus;
  priority?: Priority;
  order_by?: string;
  limit?: number;
  offset?: number;
  include?: string; // comma-separated: creator,assignee,user_stories,comments
}

export interface UserStoryListParams {
  epic_id?: string;
  creator_id?: string;
  assignee_id?: string;
  status?: UserStoryStatus;
  priority?: Priority;
  order_by?: string;
  limit?: number;
  offset?: number;
  include?: string; // comma-separated: epic,creator,assignee,acceptance_criteria,requirements,comments
}

export interface AcceptanceCriteriaListParams {
  user_story_id?: string;
  author_id?: string;
  order_by?: string;
  limit?: number;
  offset?: number;
}

export interface RequirementListParams {
  user_story_id?: string;
  acceptance_criteria_id?: string;
  type_id?: string;
  creator_id?: string;
  assignee_id?: string;
  status?: RequirementStatus;
  priority?: Priority;
  order_by?: string;
  limit?: number;
  offset?: number;
  include?: string;
}

export interface CommentListParams {
  entity_type?: EntityType;
  entity_id?: string;
  author_id?: string;
  is_resolved?: boolean;
  order_by?: string;
  limit?: number;
  offset?: number;
}

// ============================================================================
// ERROR TYPES
// ============================================================================

export interface ErrorResponse {
  error: {
    code: string;
    message: string;
  };
}

export type ErrorCode = 
  | 'VALIDATION_ERROR'
  | 'AUTHENTICATION_REQUIRED'
  | 'INSUFFICIENT_PERMISSIONS'
  | 'ENTITY_NOT_FOUND'
  | 'DELETION_CONFLICT'
  | 'INTERNAL_ERROR';

// ============================================================================
// API CLIENT CONFIGURATION
// ============================================================================

export interface ApiClientConfig {
  baseUrl: string;
  timeout?: number;
  defaultHeaders?: Record<string, string>;
}

export interface AuthTokens {
  accessToken: string;
  expiresAt: string;
}

// ============================================================================
// UTILITY TYPES
// ============================================================================

export type EntityId = string; // UUID or reference ID
export type OptionalExcept<T, K extends keyof T> = Partial<T> & Pick<T, K>;
export type RequiredFields<T, K extends keyof T> = T & Required<Pick<T, K>>;

// Helper type for API endpoints that accept either UUID or reference ID
export type EntityIdentifier = string;

// Helper type for include parameters
export type IncludeParam<T extends string> = T | `${T},${string}` | string;

// Common include options for each entity type
export type EpicInclude = 'creator' | 'assignee' | 'user_stories' | 'comments';
export type UserStoryInclude = 'epic' | 'creator' | 'assignee' | 'acceptance_criteria' | 'requirements' | 'comments';
export type AcceptanceCriteriaInclude = 'user_story' | 'author' | 'requirements' | 'comments';
export type RequirementInclude = 'user_story' | 'acceptance_criteria' | 'type' | 'creator' | 'assignee' | 'source_relationships' | 'target_relationships' | 'comments';

// ============================================================================
// CONSTANTS
// ============================================================================

export const ENTITY_TYPES = ['epic', 'user_story', 'acceptance_criteria', 'requirement'] as const;
export const USER_ROLES = ['Administrator', 'User', 'Commenter'] as const;
export const PRIORITIES = [1, 2, 3, 4] as const;
export const EPIC_STATUSES = ['Backlog', 'Draft', 'In Progress', 'Done', 'Cancelled'] as const;
export const USER_STORY_STATUSES = ['Backlog', 'Draft', 'In Progress', 'Done', 'Cancelled'] as const;
export const REQUIREMENT_STATUSES = ['Draft', 'Active', 'Obsolete'] as const;

export const PRIORITY_LABELS = {
  1: 'Critical',
  2: 'High', 
  3: 'Medium',
  4: 'Low'
} as const;

export const DEFAULT_PAGE_SIZE = 50;
export const MAX_PAGE_SIZE = 100;

// ============================================================================
// TYPE GUARDS
// ============================================================================

export function isEntityType(value: string): value is EntityType {
  return ENTITY_TYPES.includes(value as EntityType);
}

export function isUserRole(value: string): value is UserRole {
  return USER_ROLES.includes(value as UserRole);
}

export function isPriority(value: number): value is Priority {
  return PRIORITIES.includes(value as Priority);
}

export function isEpicStatus(value: string): value is EpicStatus {
  return EPIC_STATUSES.includes(value as EpicStatus);
}

export function isUserStoryStatus(value: string): value is UserStoryStatus {
  return USER_STORY_STATUSES.includes(value as UserStoryStatus);
}

export function isRequirementStatus(value: string): value is RequirementStatus {
  return REQUIREMENT_STATUSES.includes(value as RequirementStatus);
}

export function isErrorResponse(response: any): response is ErrorResponse {
  return response && typeof response === 'object' && 'error' in response;
}

export function isApiResponse<T>(response: any): response is ApiResponse<T> {
  return response && typeof response === 'object' && ('data' in response || 'error' in response);
}