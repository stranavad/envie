import { useQuery } from '@tanstack/vue-query';
import { computed, type MaybeRef, toValue } from 'vue';
import { ProjectService } from '@/services/project.service';
import { queryKeys } from './keys';

export function useProjects() {
    return useQuery({
        queryKey: queryKeys.projects,
        queryFn: () => ProjectService.getProjects(),
    });
}

export function useProject(id: MaybeRef<string>) {
    return useQuery({
        queryKey: computed(() => queryKeys.project(toValue(id))),
        queryFn: () => ProjectService.getProject(toValue(id)),
        enabled: computed(() => !!toValue(id)),
    });
}

export function useProjectConfig(projectId: MaybeRef<string>) {
    return useQuery({
        queryKey: computed(() => queryKeys.projectConfig(toValue(projectId))),
        queryFn: () => ProjectService.getConfig(toValue(projectId)),
        enabled: computed(() => !!toValue(projectId)),
    });
}

export function useProjectTeams(projectId: MaybeRef<string>) {
    return useQuery({
        queryKey: computed(() => queryKeys.projectTeams(toValue(projectId))),
        queryFn: () => ProjectService.getProjectTeams(toValue(projectId)),
        enabled: computed(() => !!toValue(projectId)),
    });
}

export function useProjectFiles(projectId: MaybeRef<string>) {
    return useQuery({
        queryKey: computed(() => queryKeys.projectFiles(toValue(projectId))),
        queryFn: () => ProjectService.getFiles(toValue(projectId)),
        enabled: computed(() => !!toValue(projectId)),
    });
}

export function useProjectTokens(projectId: MaybeRef<string>) {
    return useQuery({
        queryKey: computed(() => queryKeys.projectTokens(toValue(projectId))),
        queryFn: () => ProjectService.getTokens(toValue(projectId)),
        enabled: computed(() => !!toValue(projectId)),
    });
}
