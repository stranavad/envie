import { useQuery } from '@tanstack/vue-query';
import { computed, type MaybeRef, toValue } from 'vue';
import { OrganizationService } from '@/services/organization.service';
import { ProjectService } from '@/services/project.service';
import { queryKeys } from './keys';

export function useOrganizations() {
    return useQuery({
        queryKey: queryKeys.organizations,
        queryFn: () => OrganizationService.getOrganizations(),
    });
}

export function useOrganization(id: MaybeRef<string>) {
    return useQuery({
        queryKey: computed(() => queryKeys.organization(toValue(id))),
        queryFn: () => OrganizationService.getOrganization(toValue(id)),
        enabled: computed(() => !!toValue(id)),
    });
}

export function useOrganizationUsers(orgId: MaybeRef<string>) {
    return useQuery({
        queryKey: computed(() => queryKeys.organizationUsers(toValue(orgId))),
        queryFn: () => OrganizationService.getOrganizationUsers(toValue(orgId)),
        enabled: computed(() => !!toValue(orgId)),
    });
}

export function useOrganizationProjects(orgId: MaybeRef<string>) {
    return useQuery({
        queryKey: computed(() => queryKeys.organizationProjects(toValue(orgId))),
        queryFn: () => ProjectService.getOrganizationProjects(toValue(orgId)),
        enabled: computed(() => !!toValue(orgId)),
    });
}
