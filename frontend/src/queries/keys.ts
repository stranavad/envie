export const queryKeys = {
    // Projects
    projects: ['projects'] as const,
    project: (id: string) => ['projects', id] as const,
    projectConfig: (projectId: string) => ['projects', projectId, 'config'] as const,
    projectTeams: (projectId: string) => ['projects', projectId, 'teams'] as const,
    projectFiles: (projectId: string) => ['projects', projectId, 'files'] as const,
    projectTokens: (projectId: string) => ['projects', projectId, 'tokens'] as const,

    // Organizations
    organizations: ['organizations'] as const,
    organization: (id: string) => ['organizations', id] as const,
    organizationUsers: (orgId: string) => ['organizations', orgId, 'users'] as const,
    organizationProjects: (orgId: string) => ['organizations', orgId, 'projects'] as const,

    // Teams
    teams: (orgId: string) => ['teams', orgId] as const,
    teamMembers: (teamId: string) => ['teams', teamId, 'members'] as const,

    // Devices
    devices: ['devices'] as const,
};
