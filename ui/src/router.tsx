import {
  createRootRoute,
  createRoute,
  createRouter,
  redirect,
} from "@tanstack/react-router";

import RootLayout from "./RootLayout";
import RotationDetail from "./RotationDetail";
import RotationsList from "./RotationsList";
import Settings from "./Settings";

const rootRoute = createRootRoute({ component: RootLayout });

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  beforeLoad: () => {
    throw redirect({ to: "/rotations" });
  },
});

const rotationsListRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/rotations",
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
  indexRoute,
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
