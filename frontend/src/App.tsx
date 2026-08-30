import {
  BrowserRouter,
  Link,
  Route,
  Routes,
} from "react-router-dom";

import { HomePage } from "./pages/HomePage";
import { IncidentDetailPage } from "./pages/IncidentDetailPage";
import { IncidentsPage } from "./pages/IncidentsPage";
import { ServicesPage } from "./pages/ServicesPage";

function App() {
  return (
    <BrowserRouter>
      <header className="app-header">
        <div className="app-header-inner">
          <Link
            className="brand"
            to="/"
          >
            CloudPulse
          </Link>

          <nav className="nav-links">
            <Link to="/services">
              Services
            </Link>

            <Link to="/incidents">
              Incidents
            </Link>
          </nav>
        </div>
      </header>

      <Routes>
        <Route
          path="/"
          element={<HomePage />}
        />

        <Route
          path="/services"
          element={<ServicesPage />}
        />

        <Route
          path="/incidents"
          element={<IncidentsPage />}
        />

        <Route
          path="/incidents/:id"
          element={<IncidentDetailPage />}
        />
      </Routes>
    </BrowserRouter>
  );
}

export default App;