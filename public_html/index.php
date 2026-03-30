<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>20i Stack - Development Environment</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .status { background: #e8f5e8; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .info-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin: 20px 0; }
        .info-box { background: #f8f9fa; padding: 15px; border-radius: 5px; }
        .info-box h3 { margin-top: 0; color: #333; }
        .links { margin: 20px 0; }
        .links a { display: inline-block; margin: 5px 10px 5px 0; padding: 8px 16px; background: #007cba; color: white; text-decoration: none; border-radius: 4px; }
        .links a:hover { background: #005a87; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 20i Stack Development Environment</h1>
            <p>Your local shared hosting replica is up and running!</p>
        </div>

        <div class="status">
            <strong>✅ Environment Status:</strong> Active and Ready
        </div>

        <div class="info-grid">
            <div class="info-box">
                <h3>🔧 PHP Information</h3>
                <p><strong>Version:</strong> <?php echo PHP_VERSION; ?></p>
                <p><strong>SAPI:</strong> <?php echo php_sapi_name(); ?></p>
                <p><strong>Server:</strong> <?php echo $_SERVER['SERVER_SOFTWARE'] ?? 'Unknown'; ?></p>
            </div>

            <div class="info-box">
                <h3>📊 System Information</h3>
                <p><strong>OS:</strong> <?php echo PHP_OS; ?></p>
                <p><strong>Memory Limit:</strong> <?php echo ini_get('memory_limit'); ?></p>
                <p><strong>Max Execution Time:</strong> <?php echo ini_get('max_execution_time'); ?>s</p>
            </div>

            <div class="info-box">
                <h3>🗄️ Database Connection</h3>
                <?php
                try {
                    $pdo = new PDO('mysql:host=mariadb;dbname=devdb', 'devuser', 'devpass');
                    echo '<p style="color: green;">✅ Database Connected Successfully</p>';
                    echo '<p><strong>Server:</strong> MariaDB</p>';
                } catch (PDOException $e) {
                    echo '<p style="color: red;">❌ Database Connection Failed</p>';
                    echo '<p>Error: ' . $e->getMessage() . '</p>';
                }
                ?>
            </div>

            <div class="info-box">
                <h3>📦 Loaded Extensions</h3>
                <?php
                $extensions = ['gd', 'mysqli', 'pdo_mysql', 'zip', 'curl', 'mbstring', 'opcache'];
                foreach ($extensions as $ext) {
                    $status = extension_loaded($ext) ? '✅' : '❌';
                    echo "<p>{$status} {$ext}</p>";
                }
                ?>
            </div>
        </div>

        <div class="links">
            <h3>🔗 Quick Links</h3>
            <a href="?phpinfo=1">PHP Info</a>
            <a href="http://localhost:8081" target="_blank">phpMyAdmin</a>
            <a href="https://github.com/20i" target="_blank">20i GitHub</a>
        </div>

        <?php if (isset($_GET['phpinfo'])): ?>
        <div style="margin-top: 30px;">
            <h3>📋 Complete PHP Information</h3>
            <div style="overflow-x: auto;">
                <?php phpinfo(); ?>
            </div>
        </div>
        <?php endif; ?>

        <footer style="text-align: center; margin-top: 40px; padding-top: 20px; border-top: 1px solid #eee; color: #666;">
            <p>20i Stack Development Environment - Ready for Development</p>
        </footer>
    </div>
</body>
</html>
