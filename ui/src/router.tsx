import {
  createRootRoute,
  createRoute,
  createRouter,
} from "@tanstack/react-router";

import RotationDetail from "./RotationDetail";
import RotationsList from "./RotationsList";
import Settings from "./Settings";

const rootRoute = createRootRoute();

const rotationsListRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: RotationsList,
});

const rotationDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/rotations/$rotationId",
  component: RotationDetail,
});

const settingsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/rotations/$rotationId/settings",
  component: Settings,
});

const routeTree = rootRoute.addChildren([
  rotationsListRoute,
  rotationDetailRoute,
  settingsRoute,
]);

export const router = createRouter({ routeTree });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
