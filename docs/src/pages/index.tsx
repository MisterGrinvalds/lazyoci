import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';

import styles from './index.module.css';

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <Heading as="h1" className="hero__title">
          {siteConfig.title}
        </Heading>
        <p className="hero__subtitle">{siteConfig.tagline}</p>
        <div className={styles.buttons}>
          <Link
            className="button button--secondary button--lg"
            to="/tutorials/getting-started">
            Get Started
          </Link>
          <Link
            className="button button--secondary button--outline button--lg"
            to="/reference/cli/">
            CLI Reference
          </Link>
        </div>
      </div>
    </header>
  );
}

type CardProps = {
  title: string;
  description: string;
  to: string;
  emoji: string;
};

function Card({title, description, to, emoji}: CardProps) {
  return (
    <div className={clsx('col col--3', styles.card)}>
      <Link to={to} className={styles.cardLink}>
        <div className={styles.cardContent}>
          <span className={styles.cardEmoji}>{emoji}</span>
          <Heading as="h3">{title}</Heading>
          <p>{description}</p>
        </div>
      </Link>
    </div>
  );
}

export default function Home(): JSX.Element {
  return (
    <Layout
      title="Terminal UI for OCI registries"
      description="Browse, explore, and pull OCI container registry artifacts from your terminal.">
      <HomepageHeader />
      <main>
        <section className={styles.cards}>
          <div className="container">
            <div className="row">
              <Card
                emoji="&#x1F4DA;"
                title="Tutorials"
                description="Step-by-step guides to get you started with lazyoci."
                to="/tutorials/"
              />
              <Card
                emoji="&#x1F527;"
                title="How-to Guides"
                description="Solve specific problems: authentication, Docker integration, and more."
                to="/guides/"
              />
              <Card
                emoji="&#x1F4D6;"
                title="Reference"
                description="CLI commands, keybindings, configuration, and environment variables."
                to="/reference/"
              />
              <Card
                emoji="&#x1F4A1;"
                title="Explanation"
                description="Understand OCI concepts, authentication architecture, and design decisions."
                to="/explanation/"
              />
            </div>
          </div>
        </section>
      </main>
    </Layout>
  );
}
