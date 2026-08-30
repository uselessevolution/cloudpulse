import { apiClient } from "./client";

import type { Incident } from "../types/incident";

export async function getIncidents(): Promise<Incident[]> {
    const response =
        await apiClient.get<Incident[]>(
            "/api/incidents/",
        );

    return response.data;
}
export interface IncidentAnalysis {
    summary: string;
    possibleCauses: string[];
    recommendedActions: string[];
}
export async function getIncidentById(
    id: number,
): Promise<Incident> {
    const response =
        await apiClient.get<Incident>(
            `/api/incidents/${id}`,
        );

    return response.data;
}
export async function generateIncidentAnalysis(
    id: number,
): Promise<IncidentAnalysis> {
    const response =
        await apiClient.post<IncidentAnalysis>(
            `/api/incidents/${id}/ai-summary`,
        );

    return response.data;
}